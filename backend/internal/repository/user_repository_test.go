package repository

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aasplit/aasplit/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestDB 创建内存 sqlite 数据库并迁移测试表（每个测试独立命名内存库）。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Group{}, &model.GroupMember{}, &model.Expense{}, &model.ExpenseShare{}, &model.Settlement{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func TestUserRepositoryCreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	tests := []struct {
		name     string
		username string
		email    string
		wantErr  bool
	}{
		{name: "first user", username: "alice", email: "alice@example.com", wantErr: false},
		{name: "duplicate username", username: "alice", email: "alice2@example.com", wantErr: true},
		{name: "second user", username: "bob", email: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &model.User{Username: tt.username, PasswordHash: "x", Nickname: tt.username, Role: "user"}
			if tt.email != "" {
				em := tt.email
				u.Email = &em
			}
			err := repo.Create(u)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("create user: %v", err)
			}
			got, err := repo.FindByUsername(tt.username)
			if err != nil {
				t.Fatalf("find by username: %v", err)
			}
			if got.Username != tt.username {
				t.Fatalf("username = %s, want %s", got.Username, tt.username)
			}
		})
	}
}

func TestUserRepositoryNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	if _, err := repo.FindByUsername("ghost"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
	if _, err := repo.FindByID(999); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestUserRepositoryUpdateFields(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	u := &model.User{Username: "carol", PasswordHash: "x", Nickname: "Carol", Role: "user"}
	if err := repo.Create(u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.UpdateFields(u.ID, map[string]interface{}{"nickname": "C.C.", "role": "admin"}); err != nil {
		t.Fatalf("update fields: %v", err)
	}
	got, err := repo.FindByID(u.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Nickname != "C.C." || string(got.Role) != "admin" {
		t.Fatalf("updated user mismatch: %+v", got)
	}
}
