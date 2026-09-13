package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
)

// AuditService 审计日志业务逻辑（service 埋点入口）。
type AuditService struct {
	auditRepo *repository.AuditRepository
	logger    *slog.Logger
}

// NewAuditService 构造审计服务。
func NewAuditService(auditRepo *repository.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{auditRepo: auditRepo, logger: logger}
}

// Record 记录一条审计日志；失败仅记录日志，不阻断主流程。
func (s *AuditService) Record(userID uint, action, resourceType, resourceID, detail, requestID string) {
	log := &model.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       truncate(detail, 1024),
		RequestID:    requestID,
	}
	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Error(fmt.Sprintf("audit record failed: action=%s err=%v", action, err))
		return
	}
	s.logger.Info(fmt.Sprintf(constants.LogAuditRecorded, log.UserID, log.Action, log.ResourceType, log.ResourceID, log.RequestID))
}

// List 分页查询审计日志（管理员）。
func (s *AuditService) List(page, pageSize int, action string, userID uint) ([]model.AuditLog, int64, error) {
	logs, total, err := s.auditRepo.List(page, pageSize, action, userID)
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return logs, total, nil
}

// ListPage 返回分页查询参数（复用 List 逻辑）。
func (s *AuditService) ListPage(query *dto.AuditQuery) ([]model.AuditLog, int64, error) {
	page, pageSize := query.Page, query.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = util.DefaultPageSize
	}
	return s.List(page, pageSize, query.Action, query.UserID)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n]) + "..."
}
