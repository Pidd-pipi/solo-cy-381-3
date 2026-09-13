package service

import (
	"testing"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
	"gorm.io/gorm"
)

// newExpenseServiceFixture 构造带群组与成员的服务夹具。
func newExpenseServiceFixture(t *testing.T) (*gorm.DB, *ExpenseService, *GroupService, uint, uint, uint, uint) {
	t.Helper()
	return newExpenseServiceFixtureDB(t, newTestDB(t))
}

// newExpenseServiceFixtureDB 在给定数据库上构造带群组与成员的服务夹具（并发测试可注入文件库）。
func newExpenseServiceFixtureDB(t *testing.T, db *gorm.DB) (*gorm.DB, *ExpenseService, *GroupService, uint, uint, uint, uint) {
	t.Helper()
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	memberRepo := repository.NewGroupMemberRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	auditSvc := NewAuditService(auditRepo, newTestLogger())
	logger := newTestLogger()

	e1 := "alice@test.com"
	e2 := "bob@test.com"
	u1 := &model.User{Username: "alice", PasswordHash: "x", Nickname: "Alice", Email: &e1, Role: constants.RoleUser}
	u2 := &model.User{Username: "bob", PasswordHash: "x", Nickname: "Bob", Email: &e2, Role: constants.RoleUser}
	u3 := &model.User{Username: "carol", PasswordHash: "x", Nickname: "Carol", Role: constants.RoleUser}
	if err := userRepo.Create(u1); err != nil {
		t.Fatalf("create u1: %v", err)
	}
	if err := userRepo.Create(u2); err != nil {
		t.Fatalf("create u2: %v", err)
	}
	if err := userRepo.Create(u3); err != nil {
		t.Fatalf("create u3: %v", err)
	}
	groupSvc := NewGroupService(db, groupRepo, memberRepo, userRepo, auditSvc, logger)
	group, err := groupSvc.Create(u1.ID, &dto.CreateGroupReq{Name: "周末聚餐", Description: "测试"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := groupSvc.InviteMember(u1.ID, group.ID, "bob"); err != nil {
		t.Fatalf("invite bob: %v", err)
	}
	if err := groupSvc.InviteMember(u1.ID, group.ID, "carol"); err != nil {
		t.Fatalf("invite carol: %v", err)
	}
	expenseSvc := NewExpenseService(db, repository.NewExpenseRepository(db), memberRepo, groupRepo, userRepo, repository.NewSettlementRepository(db), auditSvc, logger)
	return db, expenseSvc, groupSvc, group.ID, u1.ID, u2.ID, u3.ID
}

func TestExpenseServiceCreateEqualSplit(t *testing.T) {
	_, svc, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixture(t)

	tests := []struct {
		name    string
		amount  float64
		payerID uint
		shares  []dto.ShareInput
		wantSum float64
		wantErr bool
		errCode int
	}{
		{name: "three way equal 300", amount: 300, payerID: aliceID, shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}}, wantSum: 300},
		{name: "odd 100.01", amount: 100.01, payerID: aliceID, shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}}, wantSum: 100.01},
		{name: "non-member payer", amount: 50, payerID: 999, shares: []dto.ShareInput{{UserID: aliceID}}, wantErr: true, errCode: 40000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.CreateExpenseReq{
				GroupID: groupID, Title: "火锅", Amount: tt.amount, Category: "dining",
				PayerID: tt.payerID, SplitType: "equal", PaidAt: "2026-08-01 12:00:00", Shares: tt.shares,
			}
			expense, err := svc.Create(aliceID, req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				ae := util.AsAppError(err)
				if ae.Code != tt.errCode {
					t.Fatalf("err code = %d, want %d", ae.Code, tt.errCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			sum := 0.0
			for _, s := range expense.Shares {
				sum += s.ShareAmount
			}
			if util.Round2(sum) != tt.wantSum {
				t.Fatalf("share sum = %.2f, want %.2f", sum, tt.wantSum)
			}
		})
	}
}

func TestExpenseServiceCreateAmountMismatch(t *testing.T) {
	_, svc, _, groupID, aliceID, bobID, _ := newExpenseServiceFixture(t)
	req := &dto.CreateExpenseReq{
		GroupID: groupID, Title: "打车", Amount: 100, Category: "transport",
		PayerID: aliceID, SplitType: "amount", PaidAt: "2026-08-02 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID, Amount: 30}, {UserID: bobID, Amount: 30}},
	}
	_, err := svc.Create(aliceID, req)
	if err == nil {
		t.Fatalf("expected mismatch error, got nil")
	}
	ae := util.AsAppError(err)
	if ae.Code != constants.CodeExpenseShareMismatch {
		t.Fatalf("err code = %d, want %d", ae.Code, constants.CodeExpenseShareMismatch)
	}
}

func TestExpenseServiceDeleteAndStatus(t *testing.T) {
	_, svc, _, groupID, aliceID, bobID, _ := newExpenseServiceFixture(t)
	req := &dto.CreateExpenseReq{
		GroupID: groupID, Title: "电影", Amount: 80, Category: "entertain",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-03 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}},
	}
	expense, err := svc.Create(aliceID, req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.Delete(aliceID, expense.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, err := svc.Get(aliceID, expense.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != constants.ExpenseRefunded {
		t.Fatalf("status = %s, want refunded", got.Status)
	}
}
