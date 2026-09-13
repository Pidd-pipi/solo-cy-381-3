package repository

import (
	"testing"
	"time"

	"github.com/aasplit/aasplit/internal/model"
)

func TestExpenseRepositoryCreateList(t *testing.T) {
	db := newTestDB(t)
	repo := NewExpenseRepository(db)
	now := time.Now()

	expense := &model.Expense{
		GroupID: 1, Title: "火锅", Amount: 300, Category: "dining", PayerID: 1,
		SplitType: "equal", PaidAt: now, Status: "active",
	}
	if err := repo.Create(nil, expense); err != nil {
		t.Fatalf("create expense: %v", err)
	}
	shares := []model.ExpenseShare{
		{ExpenseID: expense.ID, UserID: 1, ShareAmount: 100, Status: "unsettled"},
		{ExpenseID: expense.ID, UserID: 2, ShareAmount: 100, Status: "unsettled"},
		{ExpenseID: expense.ID, UserID: 3, ShareAmount: 100, Status: "unsettled"},
	}
	if err := repo.CreateShares(nil, shares); err != nil {
		t.Fatalf("create shares: %v", err)
	}

	tests := []struct {
		name     string
		params   ExpenseQueryParams
		wantList int
		wantTot  int64
	}{
		{"all", ExpenseQueryParams{GroupID: 1, Page: 1, PageSize: 10}, 1, 1},
		{"by category", ExpenseQueryParams{GroupID: 1, Category: "dining", Page: 1, PageSize: 10}, 1, 1},
		{"miss category", ExpenseQueryParams{GroupID: 1, Category: "transport", Page: 1, PageSize: 10}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, total, err := repo.List(tt.params)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(list) != tt.wantList || total != tt.wantTot {
				t.Fatalf("list=%d total=%d, want %d/%d", len(list), total, tt.wantList, tt.wantTot)
			}
		})
	}

	got, err := repo.FindByID(expense.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if len(got.Shares) != 3 {
		t.Fatalf("shares = %d, want 3", len(got.Shares))
	}
}

func TestExpenseRepositoryDeleteShares(t *testing.T) {
	db := newTestDB(t)
	repo := NewExpenseRepository(db)
	e := &model.Expense{GroupID: 1, Title: "打车", Amount: 60, Category: "transport", PayerID: 1, SplitType: "equal", PaidAt: time.Now(), Status: "active"}
	if err := repo.Create(nil, e); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.CreateShares(nil, []model.ExpenseShare{{ExpenseID: e.ID, UserID: 1, ShareAmount: 30, Status: "unsettled"}}); err != nil {
		t.Fatalf("create shares: %v", err)
	}
	if err := repo.DeleteShares(nil, e.ID); err != nil {
		t.Fatalf("delete shares: %v", err)
	}
	var n int64
	db.Model(&model.ExpenseShare{}).Where("expense_id = ?", e.ID).Count(&n)
	if n != 0 {
		t.Fatalf("remaining shares = %d, want 0", n)
	}
}
