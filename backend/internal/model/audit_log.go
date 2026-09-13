package model

import (
	"time"
)

// AuditLog 操作审计日志实体：记录关键业务动作。
type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	Action       string    `gorm:"size:64;not null;index" json:"action"`
	ResourceType string    `gorm:"size:64;not null" json:"resource_type"`
	ResourceID   string    `gorm:"size:64" json:"resource_id"`
	Detail       string    `gorm:"size:1024" json:"detail"`
	IP           string    `gorm:"size:64" json:"ip"`
	RequestID    string    `gorm:"size:64" json:"request_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名。
func (AuditLog) TableName() string { return "audit_logs" }
