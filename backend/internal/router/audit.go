package router

import (
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RegisterAuditRoutes 注册审计日志路由（管理员）。
func RegisterAuditRoutes(r *gin.RouterGroup, h *handler.AuditHandler, jwt *util.JWTManager) {
	audit := r.Group("/audit-logs")
	audit.Use(middleware.Auth(jwt), middleware.RequireRoles("admin"))
	{
		audit.GET("", h.List)
	}
}
