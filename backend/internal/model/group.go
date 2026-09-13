package model

import (
	"time"

	"github.com/aasplit/aasplit/internal/constants"
)

// Group 分账群组实体：独立账单空间，支持名称与描述编辑、归档。
type Group struct {
	ID          uint                  `gorm:"primaryKey" json:"id"`
	Name        string                `gorm:"size:128;not null" json:"name"`
	Description string                `gorm:"size:512" json:"description"`
	OwnerID     uint                  `gorm:"index;not null" json:"owner_id"`
	Status      constants.GroupStatus `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// TableName 指定表名。
func (Group) TableName() string { return "groups" }

// IsActive 判断群组是否进行中。
func (g *Group) IsActive() bool { return g.Status == constants.GroupActive }
