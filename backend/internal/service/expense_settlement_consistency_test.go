package service

import (
	"testing"

	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/repository"
	"gorm.io/gorm"
)

// buildSettlementService 在消费夹具基础上装配结算服务。
func buildSettlementService(db *gorm.DB) *SettlementService {
	return NewSettlementService(
		db,
		repository.NewSettlementRepository(db),
		repository.NewExpenseShareRepository(db),
		repository.NewGroupMemberRepository(db),
		repository.NewGroupRepository(db),
		repository.NewUserRepository(db),
		NewAuditService(repository.NewAuditRepository(db), newTestLogger()),
		newTestLogger(),
	)
}

// groupSettlementCount 统计群组内某状态的结算建议数量。
func groupSettlementCount(t *testing.T, db *gorm.DB, groupID uint, status string) int64 {
	t.Helper()
	var n int64
	if err := db.Table("settlements").Where("group_id = ? AND status = ?", groupID, status).Count(&n).Error; err != nil {
		t.Fatalf("count settlements: %v", err)
	}
	return n
}

// TestSettlementInvalidatedOnExpenseUpdate 修改消费后，旧的待结算建议失效；重新生成恢复。
func TestSettlementInvalidatedOnExpenseUpdate(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)

	if _, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID)); err != nil {
		t.Fatalf("create expense: %v", err)
	}
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("transfers = %d, want 2", len(items))
	}

	// 修改消费金额（300 三人均摊 → 600 三人均摊）。
	listed, _, err := expenseSvc.List(aliceID, groupID, &dto.ExpenseQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list expense: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expenses = %d, want 1", len(listed))
	}
	if err := expenseSvc.Update(aliceID, listed[0].ID, &dto.UpdateExpenseReq{
		Title: "火锅-改", Amount: 600, Category: "dining",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-01 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}},
	}); err != nil {
		t.Fatalf("update expense: %v", err)
	}

	// 群组列表与待结算提醒都不应再返回按旧金额生成的建议。
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending settlements after update = %d, want 0", got)
	}
	groupItems, err := settleSvc.ListByGroup(aliceID, groupID)
	if err != nil {
		t.Fatalf("list by group: %v", err)
	}
	if len(groupItems) != 0 {
		t.Fatalf("stale suggestions after update = %d, want 0", len(groupItems))
	}
	pending, err := settleSvc.ListPending(bobID)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("stale pending reminders = %d, want 0", len(pending))
	}

	// 重新生成后按新金额恢复：Bob/Carol 各应付 200，共 2 笔、合计 400。
	regenerated, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(regenerated) != 2 {
		t.Fatalf("transfers after regenerate = %d, want 2", len(regenerated))
	}
	total := 0.0
	for _, it := range regenerated {
		if it.Status != "pending" {
			t.Fatalf("status = %s, want pending", it.Status)
		}
		total += it.Amount
	}
	if total < 399.99 || total > 400.01 {
		t.Fatalf("regenerated total = %.2f, want ~400 (new amount)", total)
	}
}

// TestSettlementInvalidatedOnExpenseRefund 退款后旧待结算建议失效。
func TestSettlementInvalidatedOnExpenseRefund(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}
	if _, err := settleSvc.Generate(aliceID, groupID); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if err := expenseSvc.Delete(aliceID, expense.ID); err != nil {
		t.Fatalf("refund: %v", err)
	}

	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending settlements after refund = %d, want 0", got)
	}
	pending, err := settleSvc.ListPending(bobID)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("stale pending reminders after refund = %d, want 0", len(pending))
	}

	// 全部消费已退款，重新生成返回空建议列表。
	regenerated, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(regenerated) != 0 {
		t.Fatalf("transfers after refund+regenerate = %d, want 0", len(regenerated))
	}
}

// TestSettlementInvalidatedOnExpenseCreate 新增消费后旧待结算建议失效。
func TestSettlementInvalidatedOnExpenseCreate(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)

	first := createExpenseReq(groupID, aliceID, bobID, carolID)
	first.Title = "火锅"
	if _, err := expenseSvc.Create(aliceID, first); err != nil {
		t.Fatalf("create first: %v", err)
	}
	if _, err := settleSvc.Generate(aliceID, groupID); err != nil {
		t.Fatalf("generate: %v", err)
	}
	second := createExpenseReq(groupID, aliceID, bobID, carolID)
	second.Title = "打车"
	second.Amount = 90
	second.PaidAt = "2026-08-02 12:00:00"
	if _, err := expenseSvc.Create(aliceID, second); err != nil {
		t.Fatalf("create second: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending settlements after second create = %d, want 0", got)
	}
}

// TestSettledHistoryKeptOnExpenseWrite 消费写入只失效待结算建议，已结算历史保留。
func TestSettledHistoryKeptOnExpenseWrite(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)
	settleSvc := buildSettlementService(db)

	if _, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID)); err != nil {
		t.Fatalf("create expense: %v", err)
	}
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := settleSvc.Settle(aliceID, &dto.SettleReq{SettlementIDs: []uint{items[0].ID, items[1].ID}}); err != nil {
		t.Fatalf("settle: %v", err)
	}
	if got := groupSettlementCount(t, db, groupID, "settled"); got != 2 {
		t.Fatalf("settled before update = %d, want 2", got)
	}

	listed, _, err := expenseSvc.List(aliceID, groupID, &dto.ExpenseQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list expense: %v", err)
	}
	if err := expenseSvc.Update(aliceID, listed[0].ID, &dto.UpdateExpenseReq{
		Title: "火锅-改", Amount: 600, Category: "dining",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-01 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}},
	}); err != nil {
		t.Fatalf("update expense: %v", err)
	}

	// 已结算记录是已发生的历史，不被消费写入联动删除。
	if got := groupSettlementCount(t, db, groupID, "settled"); got != 2 {
		t.Fatalf("settled history after update = %d, want 2", got)
	}
	if got := groupSettlementCount(t, db, groupID, "pending"); got != 0 {
		t.Fatalf("pending after update = %d, want 0", got)
	}
}
