package repository

import (
	"testing"

	"github.com/aasplit/aasplit/internal/model"
)

func TestGroupRepositoryListByUser(t *testing.T) {
	db := newTestDB(t)
	groupRepo := NewGroupRepository(db)
	memberRepo := NewGroupMemberRepository(db)

	g := &model.Group{Name: "周末聚餐", Description: "好友聚餐", OwnerID: 1, Status: "active"}
	if err := groupRepo.Create(g); err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := memberRepo.Create(&model.GroupMember{GroupID: g.ID, UserID: 1, Role: "owner", Status: "active"}); err != nil {
		t.Fatalf("create member: %v", err)
	}
	if err := memberRepo.Create(&model.GroupMember{GroupID: g.ID, UserID: 2, Role: "normal", Status: "active"}); err != nil {
		t.Fatalf("create member2: %v", err)
	}

	groups, total, err := groupRepo.ListByUser(1, 1, 10)
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if total != 1 || len(groups) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1", total, len(groups))
	}
	count, err := memberRepo.Count(g.ID)
	if err != nil {
		t.Fatalf("count members: %v", err)
	}
	if count != 2 {
		t.Fatalf("member count = %d, want 2", count)
	}
}

func TestGroupRepositoryLockByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupRepository(db)
	g := &model.Group{Name: "旅行分摊", OwnerID: 1, Status: "active"}
	if err := repo.Create(g); err != nil {
		t.Fatalf("create: %v", err)
	}
	locked, err := repo.LockByID(g.ID)
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	if locked.ID != g.ID {
		t.Fatalf("locked id = %d, want %d", locked.ID, g.ID)
	}
}
