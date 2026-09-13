package dto

import "github.com/aasplit/aasplit/internal/constants"

// RegisterReq 注册请求。
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Nickname string `json:"nickname" binding:"required,min=1,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=128"`
}

// LoginReq 登录请求。
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginResp 登录响应。
type LoginResp struct {
	Token string    `json:"token"`
	User  *UserResp `json:"user"`
}

// UserResp 用户响应。
type UserResp struct {
	ID        uint               `json:"id"`
	Username  string             `json:"username"`
	Nickname  string             `json:"nickname"`
	Email     string             `json:"email"`
	Avatar    string             `json:"avatar"`
	Role      constants.UserRole `json:"role"`
	Status    string             `json:"status"`
	CreatedAt string             `json:"created_at"`
}

// UpdateProfileReq 更新个人资料请求。
type UpdateProfileReq struct {
	Nickname string `json:"nickname" binding:"required,min=1,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=128"`
	Avatar   string `json:"avatar" binding:"omitempty,max=512"`
}

// UpdateRoleReq 管理员修改用户角色请求。
type UpdateRoleReq struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

// ChangePasswordReq 修改密码请求。
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=72"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}
