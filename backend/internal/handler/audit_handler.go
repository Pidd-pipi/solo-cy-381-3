package handler

import (
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志 HTTP 处理器。
type AuditHandler struct {
	auditSvc *service.AuditService
	userRepo *repository.UserRepository
}

// NewAuditHandler 构造审计处理器。
func NewAuditHandler(auditSvc *service.AuditService, userRepo *repository.UserRepository) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc, userRepo: userRepo}
}

// List 分页查询审计日志（管理员）。
func (h *AuditHandler) List(c *gin.Context) {
	var query dto.AuditQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	logs, total, err := h.auditSvc.ListPage(&query)
	if err != nil {
		util.Fail(c, err)
		return
	}
	names := h.usernameMap(logs)
	list := make([]dto.AuditLogResp, 0, len(logs))
	for i := range logs {
		username := "system"
		if u, ok := names[logs[i].UserID]; ok {
			username = u
		}
		list = append(list, dto.AuditLogResp{
			ID:           logs[i].ID,
			UserID:       logs[i].UserID,
			Username:     username,
			Action:       logs[i].Action,
			ResourceType: logs[i].ResourceType,
			ResourceID:   logs[i].ResourceID,
			Detail:       logs[i].Detail,
			IP:           logs[i].IP,
			RequestID:    logs[i].RequestID,
			CreatedAt:    util.FormatDateTime(logs[i].CreatedAt),
		})
	}
	page, pageSize := util.ParsePagination(c)
	util.OK(c, util.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// usernameMap 批量查询日志涉及用户的用户名。
func (h *AuditHandler) usernameMap(logs []model.AuditLog) map[uint]string {
	ids := make([]uint, 0, len(logs))
	seen := make(map[uint]struct{}, len(logs))
	for _, l := range logs {
		if _, ok := seen[l.UserID]; ok {
			continue
		}
		seen[l.UserID] = struct{}{}
		if l.UserID > 0 {
			ids = append(ids, l.UserID)
		}
	}
	result := make(map[uint]string, len(ids))
	for _, id := range ids {
		if u, err := h.userRepo.FindByID(id); err == nil {
			result[id] = u.Username
		}
	}
	return result
}
