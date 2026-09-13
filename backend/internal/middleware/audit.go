package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// Audit 通用 HTTP 审计中间件：记录所有写操作（POST/PUT/DELETE）到审计日志表。
func Audit(auditRepo *repository.AuditRepository, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			return
		}
		status := c.Writer.Status()
		if status >= 400 {
			return
		}
		userID := util.GetUserID(c)
		action := strings.ToLower(method) + ":" + strings.ToLower(strings.TrimPrefix(c.FullPath(), "/api/v1"))
		log := &model.AuditLog{
			UserID:       userID,
			Action:       action,
			ResourceType: resourceTypeFromPath(c.FullPath()),
			ResourceID:   c.Param("id"),
			Detail:       fmt.Sprintf("%s %s", method, c.Request.URL.Path),
			IP:           util.GetClientIP(c),
			RequestID:    util.GetRequestID(c),
		}
		if err := auditRepo.Create(log); err != nil {
			logger.Error(fmt.Sprintf("http audit failed: path=%s err=%v", c.Request.URL.Path, err))
			return
		}
		logger.Info(fmt.Sprintf(constants.LogAuditRecorded, log.UserID, log.Action, log.ResourceType, log.ResourceID, log.RequestID))
	}
}

// resourceTypeFromPath 从路由路径推断资源类型。
func resourceTypeFromPath(path string) string {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for i, s := range segs {
		if s == "api" && i+2 < len(segs) {
			return segs[i+2]
		}
	}
	if len(segs) > 0 {
		return segs[len(segs)-1]
	}
	return "unknown"
}
