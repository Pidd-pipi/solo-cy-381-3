package service

import (
	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// migrateAll 迁移全部测试表。
func migrateAll(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.GroupMember{},
		&model.Expense{},
		&model.ExpenseShare{},
		&model.Settlement{},
		&model.AuditLog{},
	)
}
