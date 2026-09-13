// Package database 负责数据库与 Redis 连接初始化。
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect 连接 PostgreSQL 并自动迁移表结构。
func Connect(dsn string, log logger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.GroupMember{},
		&model.Expense{},
		&model.ExpenseShare{},
		&model.Settlement{},
		&model.AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return db, nil
}

// ConnectRedis 连接 Redis，用于限流中间件。
func ConnectRedis(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: 0})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}

// EnsureBootstrapUser 初始化默认管理员账号（幂等）。
func EnsureBootstrapUser(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return fmt.Errorf("count admin: %w", err)
	}
	if count > 0 {
		return nil
	}
	admin := &model.User{
		Username: "admin",
		Nickname: "系统管理员",
		Role:     constants.RoleAdmin,
		Status:   model.UserStatusActive,
		Avatar:   "",
	}
	adminEmail := "admin@aasplit.local"
	admin.Email = &adminEmail
	// 使用初始化密码 admin123 生成真实哈希（开发环境）。
	hash, err := HashPasswordBootstrap("admin123")
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	admin.PasswordHash = hash
	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}
