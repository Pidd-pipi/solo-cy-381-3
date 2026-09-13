package service

import (
	"testing"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/util"
)

// 本文件聚焦“消费写入后结算建议失效”的边界场景，不改动任何生产代码：
//  1. 没有旧建议时创建/修改/退款（失效操作为无副作用的空操作）；
//  2. 同一群组存在多笔消费时修改其中一笔/退款其中一笔；
//  3. 群组内只有已结算历史、没有待结算建议；
//  4. 重复退款。
//
// 每个用例都验证：操作完成后过期建议不再从群组列表 / 待结算提醒返回，
// 且重新生成后结果恢复正常。

// sumTransferAmounts 汇总结算建议的转账金额。
func sumTransferAmounts(items []model.Settlement) float64 {
	total := 0.0
	for _, it := range items {
		total += it.Amount
	}
	return total
}

// settleIDs 收集结算建议 ID。
func settleIDs(items []model.Settlement) []uint {
	ids := make([]uint, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	return ids
}

// assertNoPendingSuggestions 断言群组列表与每个成员的待结算提醒都不含待结算建议。
func assertNoPendingSuggestions(t *testing.T, settleSvc *SettlementService, groupID uint, memberIDs ...uint) {
	t.Helper()
	for _, uid := range memberIDs {
		groupItems, err := settleSvc.ListByGroup(uid, groupID)
		if err != nil {
			t.Fatalf("list by group as user %d: %v", uid, err)
		}
		for _, it := range groupItems {
			if it.Status == constants.SettlementPending {
				t.Fatalf("group list returned stale pending suggestion %d for user %d", it.ID, uid)
			}
		}
		pending, err := settleSvc.ListPending(uid)
		if err != nil {
			t.Fatalf("list pending for user %d: %v", uid, err)
		}
		if len(pending) != 0 {
			t.Fatalf("user %d has %d stale pending reminders, want 0", uid, len(pending))
		}
	}
}

// updateReq 构造与 createExpenseReq 对应的均摊更新请求（Alice 付款，三人参与）。
func updateReq(title string, amount float64, paidAt string, participantIDs ...uint) *dto.UpdateExpenseReq {
	shares := make([]dto.ShareInput, 0, len(participantIDs))
	for _, id := range participantIDs {
		shares = append(shares, dto.ShareInput{UserID: id})
	}
	return &dto.UpdateExpenseReq{
		Title: title, Amount: amount, Category: "dining",
		PayerID: participantIDs[0], SplitType: "equal", PaidAt: paidAt, Shares: shares,
	}
}

// firstExpenseID 列出群组内有效消费，断言只有一条并返回其 ID。
func firstExpenseID(t *testing.T, expenseSvc *ExpenseService, groupID, userID uint) uint {
	t.Helper()
	listed, _, err := expenseSvc.List(userID, groupID, &dto.ExpenseQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list expense: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("active expenses = %d, want 1", len(listed))
	}
	return listed[0].ID
}

// TestCreateWithoutPriorSuggestions 从未生成过建议时新增消费：不报错、不产生建议；首次生成即正常。
func TestCreateWithoutPriorSuggestions(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)
	memberIDs := []uint{aliceID, bobID, carolID}

	if _, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID)); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending = %d, want 0 (never generated)", got)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)

	// 首次生成即正常：300 三人均摊，Bob/Carol 各应付 100，共 2 笔合计 200。
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("transfers = %d, want 2", len(items))
	}
	if total := sumTransferAmounts(items); total < 199.99 || total > 200.01 {
		t.Fatalf("generated total = %.2f, want ~200", total)
	}
}

// TestUpdateWithoutPriorSuggestions 从未生成过建议时修改消费：失效操作空转；随后生成按新金额。
func TestUpdateWithoutPriorSuggestions(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)
	memberIDs := []uint{aliceID, bobID, carolID}

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := expenseSvc.Update(aliceID, expense.ID,
		updateReq("火锅-改", 600, "2026-08-01 12:00:00", memberIDs...)); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending = %d, want 0 (never generated)", got)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)

	// 重新（首次）生成按修改后金额：600 三人均摊，各应付 200，合计 400。
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if total := sumTransferAmounts(items); total < 399.99 || total > 400.01 {
		t.Fatalf("generated total = %.2f, want ~400 (updated amount)", total)
	}
}

// TestRefundWithoutPriorSuggestions 从未生成过建议时退款：失效操作空转；随后生成为空。
func TestRefundWithoutPriorSuggestions(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)
	memberIDs := []uint{aliceID, bobID, carolID}

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := expenseSvc.Delete(aliceID, expense.ID); err != nil {
		t.Fatalf("refund: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending = %d, want 0 (never generated)", got)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)

	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("generated after refund = %d transfers, want 0", len(items))
	}
}

// TestSettlementInvalidatedWithMultipleExpenses 同一群组多笔消费：修改/退款任一笔都失效旧建议，重算覆盖全部有效消费。
func TestSettlementInvalidatedWithMultipleExpenses(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	memberIDs := []uint{aliceID, bobID, carolID}

	// 两笔 Alice 垫付的三人均摊消费：300 + 90 = 390，各人应付 130，Alice 应收 260。
	first := createExpenseReq(groupID, aliceID, bobID, carolID)
	if _, err := expenseSvc.Create(aliceID, first); err != nil {
		t.Fatalf("create first: %v", err)
	}
	second := createExpenseReq(groupID, aliceID, bobID, carolID)
	second.Title = "打车"
	second.Amount = 90
	second.PaidAt = "2026-08-02 12:00:00"
	if _, err := expenseSvc.Create(aliceID, second); err != nil {
		t.Fatalf("create second: %v", err)
	}

	// 先验证“修改其中一笔”：重建一个结算服务读取同一数据库。
	settleSvc := buildSettlementService(db)
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if total := sumTransferAmounts(items); total < 259.99 || total > 260.01 {
		t.Fatalf("initial total = %.2f, want ~260 (390-130)", total)
	}

	// List 按 paid_at DESC，第二笔（8-02）在前，第一笔（8-01）在后。
	listed, _, err := expenseSvc.List(aliceID, groupID, &dto.ExpenseQuery{Page: 1, PageSize: 10})
	if err != nil || len(listed) != 2 {
		t.Fatalf("list expenses: %v len=%d", err, len(listed))
	}
	target := listed[1]

	if err := expenseSvc.Update(aliceID, target.ID,
		updateReq("火锅-改", 600, "2026-08-01 12:00:00", memberIDs...)); err != nil {
		t.Fatalf("update: %v", err)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)
	// 第一笔 300 -> 600：总额 690，各人应付 230，Alice 应收 460。
	regenerated, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate after update: %v", err)
	}
	if len(regenerated) != 2 {
		t.Fatalf("transfers after update = %d, want 2", len(regenerated))
	}
	if total := sumTransferAmounts(regenerated); total < 459.99 || total > 460.01 {
		t.Fatalf("regenerated total after update = %.2f, want ~460 (690-230)", total)
	}

	// 再退掉被改成 600 的第一笔：仅剩 90 的第二笔，各人应付 30，Alice 应收 60。
	if err := expenseSvc.Delete(aliceID, target.ID); err != nil {
		t.Fatalf("refund: %v", err)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)
	regenerated, err = settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate after refund: %v", err)
	}
	if len(regenerated) != 2 {
		t.Fatalf("transfers after refund = %d, want 2", len(regenerated))
	}
	if total := sumTransferAmounts(regenerated); total < 59.99 || total > 60.01 {
		t.Fatalf("regenerated total after refund = %.2f, want ~60 (90-30)", total)
	}
	active, _, err := expenseSvc.List(aliceID, groupID, &dto.ExpenseQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active expenses after refund = %d, want 1", len(active))
	}
}

// TestSettlementOnlySettledHistoryOnUpdate 只有已结算历史时修改消费：历史保留、无待结算建议；重新生成按新金额恢复。
func TestSettlementOnlySettledHistoryOnUpdate(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)
	memberIDs := []uint{aliceID, bobID, carolID}

	if _, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID)); err != nil {
		t.Fatalf("create: %v", err)
	}
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := settleSvc.Settle(aliceID, &dto.SettleReq{SettlementIDs: settleIDs(items)}); err != nil {
		t.Fatalf("settle: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "settled"); got != 2 {
		t.Fatalf("settled history = %d, want 2", got)
	}

	expenseID := firstExpenseID(t, expenseSvc, groupID, aliceID)
	if err := expenseSvc.Update(aliceID, expenseID,
		updateReq("火锅-改", 600, "2026-08-01 12:00:00", memberIDs...)); err != nil {
		t.Fatalf("update: %v", err)
	}
	// 修改联动只删 pending，已结算历史原样保留；群组列表/提醒都不再出现待结算建议。
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)
	if got := groupSettlementCount(t, db, groupID, "settled"); got != 2 {
		t.Fatalf("settled history after update = %d, want 2", got)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending after update = %d, want 0", got)
	}

	// 重新生成后按新金额恢复（生成结果全部为新的 pending）。
	regenerated, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(regenerated) != 2 {
		t.Fatalf("regenerated transfers = %d, want 2", len(regenerated))
	}
	if total := sumTransferAmounts(regenerated); total < 399.99 || total > 400.01 {
		t.Fatalf("regenerated total = %.2f, want ~400 (updated amount)", total)
	}
	for _, it := range regenerated {
		if it.Status != constants.SettlementPending {
			t.Fatalf("regenerated suggestion %d status = %s, want pending", it.ID, it.Status)
		}
	}
}

// TestSettlementOnlySettledHistoryOnRefund 只有已结算历史时退款：历史保留、无待结算建议；重新生成为空。
func TestSettlementOnlySettledHistoryOnRefund(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)
	memberIDs := []uint{aliceID, bobID, carolID}

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := settleSvc.Settle(aliceID, &dto.SettleReq{SettlementIDs: settleIDs(items)}); err != nil {
		t.Fatalf("settle: %v", err)
	}

	if err := expenseSvc.Delete(aliceID, expense.ID); err != nil {
		t.Fatalf("refund: %v", err)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)
	if got := groupSettlementCount(t, db, groupID, "settled"); got != 2 {
		t.Fatalf("settled history after refund = %d, want 2", got)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending after refund = %d, want 0", got)
	}

	regenerated, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(regenerated) != 0 {
		t.Fatalf("regenerated after refund = %d transfers, want 0", len(regenerated))
	}
}

// TestSettlementInvalidatedOnRepeatedRefund 重复退款返回冲突错误且不改变已失效的建议状态。
func TestSettlementInvalidatedOnRepeatedRefund(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)
	memberIDs := []uint{aliceID, bobID, carolID}

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := settleSvc.Generate(aliceID, groupID); err != nil {
		t.Fatalf("generate: %v", err)
	}

	if err := expenseSvc.Delete(aliceID, expense.ID); err != nil {
		t.Fatalf("first refund: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending after first refund = %d, want 0", got)
	}
	got, err := expenseSvc.Get(aliceID, expense.ID)
	if err != nil {
		t.Fatalf("get refunded: %v", err)
	}
	if got.Status != constants.ExpenseRefunded {
		t.Fatalf("status = %s, want refunded", got.Status)
	}

	//再次退款：返回冲突错误码（既有校验规则不变）。
	err = expenseSvc.Delete(aliceID, expense.ID)
	if err == nil {
		t.Fatalf("second refund: expected conflict error, got nil")
	}
	ae := util.AsAppError(err)
	if ae == nil || ae.Code != constants.CodeConflict {
		t.Fatalf("second refund err = %v, want code %d", err, constants.CodeConflict)
	}

	// 重复退款无副作用：记录仍为 refunded，群组列表与待结算提醒均无建议。
	got, err = expenseSvc.Get(aliceID, expense.ID)
	if err != nil {
		t.Fatalf("get after second refund: %v", err)
	}
	if got.Status != constants.ExpenseRefunded {
		t.Fatalf("status after second refund = %s, want refunded", got.Status)
	}
	assertNoPendingSuggestions(t, settleSvc, groupID, memberIDs...)

	// 全部消费已退款，重新生成恢复为“无待结算建议”。
	regenerated, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(regenerated) != 0 {
		t.Fatalf("regenerated = %d transfers, want 0", len(regenerated))
	}

	// 重新生成后再重复退款，仍返回同样的冲突错误码。
	err = expenseSvc.Delete(aliceID, expense.ID)
	if err == nil {
		t.Fatalf("third refund: expected conflict error, got nil")
	}
	ae = util.AsAppError(err)
	if ae == nil || ae.Code != constants.CodeConflict {
		t.Fatalf("third refund err = %v, want code %d", err, constants.CodeConflict)
	}
}
