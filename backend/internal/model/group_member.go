package model

import (
	"time"

	"github.com/aasplit/aasplit/internal/constants"
)

// MemberRole 群组内角色。
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"  // 群主
	MemberRoleNormal MemberRole = "normal" // 普通成员
)

// GroupMember 群组成员实体：成员邀请与群组归属。
type GroupMember struct {
	ID        uint                  `gorm:"primaryKey" json:"id"`
	GroupID   uint                  `gorm:"index:idx_group_user,unique;not null" json:"group_id"`
	UserID    uint                  `gorm:"index:idx_group_user,unique;not null" json:"user_id"`
	Role      MemberRole            `gorm:"size:16;not null;default:normal" json:"role"`
	InvitedBy uint                  `json:"invited_by"`
	Status    constants.GroupStatus `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名。
func (GroupMember) TableName() string { return "group_members" }
