package router

import (
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RegisterSettlementRoutes 注册结算建议路由。
func RegisterSettlementRoutes(r *gin.RouterGroup, h *handler.SettlementHandler, jwt *util.JWTManager) {
	groups := r.Group("/groups/:id")
	groups.Use(middleware.Auth(jwt))
	{
		groups.POST("/settlements/generate", h.Generate)
		groups.GET("/settlements", h.ListByGroup)
		groups.GET("/balances", h.Balances)
	}
	settlements := r.Group("/settlements")
	settlements.Use(middleware.Auth(jwt))
	{
		settlements.GET("/pending", h.ListPending)
		settlements.POST("/settle", h.Settle)
	}
}
