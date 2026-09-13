package service

import (
	"testing"

	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/repository"
)

func TestSettlementServiceGenerate(t *testing.T) {
	db, expenseSvc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)

	// Alice 付 300 三人均摊 → Bob 应付 100，Carol 应付 100
	req := &dto.CreateExpenseReq{
		GroupID: groupID, Title: "火锅", Amount: 300, Category: "dining",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-01 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}},
	}
	if _, err := expenseSvc.Create(aliceID, req); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	settleRepo := repository.NewSettlementRepository(db)
	shareRepo := repository.NewExpenseShareRepository(db)
	memberRepo := repository.NewGroupMemberRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	userRepo := repository.NewUserRepository(db)
	auditSvc := NewAuditService(repository.NewAuditRepository(db), newTestLogger())
	svc := NewSettlementService(db, settleRepo, shareRepo, memberRepo, groupRepo, userRepo, auditSvc, newTestLogger())

	items, err := svc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	total := 0.0
	for _, it := range items {
		total += it.Amount
		if it.Status != "pending" {
			t.Fatalf("status = %s, want pending", it.Status)
		}
	}
	if len(items) != 2 {
		t.Fatalf("transfers = %d, want 2", len(items))
	}
	if total < 199.99 || total > 200.01 {
		t.Fatalf("transfer total = %.2f, want ~200", total)
	}

	affected, err := svc.Settle(aliceID, &dto.SettleReq{SettlementIDs: []uint{items[0].ID, items[1].ID}})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if affected != 2 {
		t.Fatalf("affected = %d, want 2", affected)
	}
	pending, err := svc.ListPending(bobID)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending after settle = %d, want 0", len(pending))
	}
}
