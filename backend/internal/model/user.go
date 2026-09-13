// Package model 定义核心实体，model 不依赖任何上层包。
package model

import (
	"time"

	"github.com/aasplit/aasplit/internal/constants"
)

// UserStatus 用户账号状态。
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 正常
	UserStatusDisabled UserStatus = "disabled" // 禁用
)

// User 用户实体：注册登录、个人头像昵称管理、角色权限。
type User struct {
	ID           uint               `gorm:"primaryKey" json:"id"`
	Username     string             `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string             `gorm:"size:255;not null" json:"-"`
	Nickname     string             `gorm:"size:64;not null" json:"nickname"`
	Email        *string            `gorm:"size:128;uniqueIndex" json:"email"`
	Avatar       string             `gorm:"size:512" json:"avatar"`
	Role         constants.UserRole `gorm:"size:16;not null;default:user" json:"role"`
	Status       UserStatus         `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// TableName 指定表名。
func (User) TableName() string { return "users" }

// IsActive 判断账号是否可用。
func (u *User) IsActive() bool { return u.Status == UserStatusActive }

// IsAdmin 判断是否管理员。
func (u *User) IsAdmin() bool { return u.Role == constants.RoleAdmin }
