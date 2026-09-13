package service

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := migrateAll(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestUserServiceRegister(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewUserService(userRepo, util.NewJWTManager("test-secret-0123456789", time.Hour), newTestLogger())

	tests := []struct {
		name    string
		req     *dto.RegisterReq
		wantErr bool
		errCode int
	}{
		{name: "valid register", req: &dto.RegisterReq{Username: "alice", Password: "secret123", Nickname: "Alice", Email: "alice@test.com"}, wantErr: false},
		{name: "duplicate username", req: &dto.RegisterReq{Username: "alice", Password: "secret123", Nickname: "Alice2"}, wantErr: true, errCode: 40901},
		{name: "duplicate email", req: &dto.RegisterReq{Username: "bob", Password: "secret123", Nickname: "Bob", Email: "alice@test.com"}, wantErr: true, errCode: 40902},
		{name: "another valid", req: &dto.RegisterReq{Username: "carol", Password: "secret123", Nickname: "Carol"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := svc.Register(tt.req)
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
				t.Fatalf("register: %v", err)
			}
			if u.Username != tt.req.Username {
				t.Fatalf("username = %s, want %s", u.Username, tt.req.Username)
			}
		})
	}
}

func TestUserServiceLogin(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewUserService(userRepo, util.NewJWTManager("test-secret-0123456789", time.Hour), newTestLogger())
	if _, err := svc.Register(&dto.RegisterReq{Username: "alice", Password: "secret123", Nickname: "Alice"}); err != nil {
		t.Fatalf("seed register: %v", err)
	}

	tests := []struct {
		name    string
		req     *dto.LoginReq
		wantErr bool
		errCode int
	}{
		{name: "correct password", req: &dto.LoginReq{Username: "alice", Password: "secret123"}, wantErr: false},
		{name: "wrong password", req: &dto.LoginReq{Username: "alice", Password: "wrongpass"}, wantErr: true, errCode: 40101},
		{name: "unknown user", req: &dto.LoginReq{Username: "ghost", Password: "secret123"}, wantErr: true, errCode: 40101},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, user, err := svc.Login(tt.req)
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
				t.Fatalf("login: %v", err)
			}
			if token == "" || user == nil {
				t.Fatalf("token/user empty")
			}
		})
	}
}
