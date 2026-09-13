package repository

import (
	"fmt"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// AuditRepository 审计日志数据访问。
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 构造审计日志仓储。
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create 记录审计日志。
func (r *AuditRepository) Create(log *model.AuditLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// List 分页查询审计日志（含用户名）。
func (r *AuditRepository) List(page, pageSize int, action string, userID uint) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	query := r.db.Model(&model.AuditLog{})
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, total, nil
}
