package model

import (
	"time"

	"github.com/aasplit/aasplit/internal/constants"
)

// Expense 消费记录实体：金额、类别、付款人、参与人分摊与小票图片。
type Expense struct {
	ID         uint                      `gorm:"primaryKey" json:"id"`
	GroupID    uint                      `gorm:"index;not null" json:"group_id"`
	Title      string                    `gorm:"size:128;not null" json:"title"`
	Amount     float64                   `gorm:"type:double precision;not null" json:"amount"`
	Category   constants.ExpenseCategory `gorm:"size:32;not null;index" json:"category"`
	PayerID    uint                      `gorm:"index;not null" json:"payer_id"`
	SplitType  constants.SplitType       `gorm:"size:16;not null" json:"split_type"`
	PaidAt     time.Time                 `gorm:"index;not null" json:"paid_at"`
	ReceiptURL string                    `gorm:"size:512" json:"receipt_url"`
	Status     constants.ExpenseStatus   `gorm:"size:16;not null;default:active" json:"status"`
	CreatedBy  uint                      `json:"created_by"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`

	// 关联（不持久化外键约束，仅用于查询填充）
	Payer  *User          `gorm:"foreignKey:PayerID" json:"payer,omitempty"`
	Shares []ExpenseShare `gorm:"foreignKey:ExpenseID" json:"shares,omitempty"`
}

// TableName 指定表名。
func (Expense) TableName() string { return "expenses" }

// IsActive 判断消费记录是否有效。
func (e *Expense) IsActive() bool { return e.Status == constants.ExpenseActive }
