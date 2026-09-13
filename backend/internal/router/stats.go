package router

import (
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RegisterStatsRoutes 注册数据统计路由。
func RegisterStatsRoutes(r *gin.RouterGroup, h *handler.StatsHandler, jwt *util.JWTManager) {
	stats := r.Group("/groups/:id/stats")
	stats.Use(middleware.Auth(jwt))
	{
		stats.GET("", h.GetGroupStats)
	}
}
