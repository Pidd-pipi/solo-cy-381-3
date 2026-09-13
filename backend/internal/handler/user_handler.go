// Package handler 处理 HTTP 请求，将 service 错误二次包装为标准 JSON 响应。
package handler

import (
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户 HTTP 处理器。
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 构造用户处理器。
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Register 注册。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	user, err := h.userSvc.Register(&req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.Created(c, toUserResp(user))
}

// Login 登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	token, user, err := h.userSvc.Login(&req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, dto.LoginResp{Token: token, User: toUserResp(user)})
}

// GetMe 查询当前用户。
func (h *UserHandler) GetMe(c *gin.Context) {
	user, err := h.userSvc.GetByID(util.GetUserID(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, toUserResp(user))
}

// UpdateProfile 更新个人资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	user, err := h.userSvc.UpdateProfile(util.GetUserID(c), &req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, toUserResp(user))
}

// ChangePassword 修改密码。
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	if err := h.userSvc.ChangePassword(util.GetUserID(c), &req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "密码修改成功"})
}

// ListUsers 管理员分页查询用户。
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, pageSize := util.ParsePagination(c)
	users, total, err := h.userSvc.List(page, pageSize)
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]*dto.UserResp, 0, len(users))
	for i := range users {
		list = append(list, toUserResp(&users[i]))
	}
	util.OK(c, util.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// ChangeRole 管理员修改角色。
func (h *UserHandler) ChangeRole(c *gin.Context) {
	var req dto.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	targetID := parseIDParam(c)
	if err := h.userSvc.ChangeRole(util.GetUserID(c), targetID, req.Role); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "角色更新成功"})
}

// toUserResp 用户模型转响应。
func toUserResp(u *model.User) *dto.UserResp {
	email := ""
	if u.Email != nil {
		email = *u.Email
	}
	return &dto.UserResp{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     email,
		Avatar:    u.Avatar,
		Role:      u.Role,
		Status:    string(u.Status),
		CreatedAt: util.FormatDateTime(u.CreatedAt),
	}
}
