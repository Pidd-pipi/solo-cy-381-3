package service

import (
	"testing"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// createExpenseReq 构造测试用的三人均摊创建请求。
func createExpenseReq(groupID, aliceID, bobID, carolID uint) *dto.CreateExpenseReq {
	return &dto.CreateExpenseReq{
		GroupID: groupID, Title: "火锅", Amount: 300, Category: "dining",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-01 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}},
	}
}

// countRows 统计表内记录数。
func countRows(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var n int64
	if err := db.Table(table).Count(&n).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// TestExpenseCreateRollback 校验：写分摊明细失败时，消费记录随同一事务一起回滚。
func TestExpenseCreateRollback(t *testing.T) {
	db, svc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)

	// 移除分摊明细表，使“写记录成功、写明细失败”这一步必然报错。
	if err := db.Migrator().DropTable("expense_shares"); err != nil {
		t.Fatalf("drop expense_shares: %v", err)
	}
	if _, err := svc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID)); err == nil {
		t.Fatalf("expected create error, got nil")
	}
	if n := countRows(t, db, "expenses"); n != 0 {
		t.Fatalf("expenses after failed create = %d, want 0 (whole tx rolled back)", n)
	}
}

// TestExpenseUpdateRollback 校验：重建分摊明细失败时，记录更新与旧明细删除一起回滚，数据保持原状。
func TestExpenseUpdateRollback(t *testing.T) {
	db, svc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)

	expense, err := svc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 更新流程会在同一事务内执行 记录更新 → 删除旧明细 → 写入新明细；让最后一步失败。
	if err := db.Migrator().DropTable("expense_shares"); err != nil {
		t.Fatalf("drop expense_shares: %v", err)
	}
	req := &dto.UpdateExpenseReq{
		Title: "改后火锅", Amount: 900, Category: "dining",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-02 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}},
	}
	if err := svc.Update(aliceID, expense.ID, req); err == nil {
		t.Fatalf("expected update error, got nil")
	}

	var got model.Expense
	if err := db.First(&got, expense.ID).Error; err != nil {
		t.Fatalf("reload expense: %v", err)
	}
	if got.Title == req.Title || got.Amount == req.Amount {
		t.Fatalf("expense update was not rolled back: title=%q amount=%.2f", got.Title, got.Amount)
	}
	if got.Status != constants.ExpenseActive {
		t.Fatalf("expense status = %s, want active", got.Status)
	}
}

// TestExpenseDeleteRollback 校验：状态更新这一步失败时退款事务不提交，消费记录保持 active。
func TestExpenseDeleteRollback(t *testing.T) {
	db, svc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)

	expense, err := svc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 在 expenses 上制造一个必然失败的 UPDATE，使 UpdateStatus 这一步报错。
	if err := db.Exec(`CREATE TRIGGER fail_expense_update BEFORE UPDATE ON expenses
BEGIN
	SELECT RAISE(ABORT, 'blocked expense update');
END;`).Error; err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	if err := svc.Delete(aliceID, expense.ID); err == nil {
		t.Fatalf("expected delete error, got nil")
	}

	var got model.Expense
	if err := db.First(&got, expense.ID).Error; err != nil {
		t.Fatalf("reload expense: %v", err)
	}
	if got.Status != constants.ExpenseActive {
		t.Fatalf("expense status = %s, want active (failed refund must roll back)", got.Status)
	}
}
