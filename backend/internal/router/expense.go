package router

import (
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RegisterExpenseRoutes 注册消费记录路由。
func RegisterExpenseRoutes(r *gin.RouterGroup, h *handler.ExpenseHandler, jwt *util.JWTManager) {
	groups := r.Group("/groups/:id/expenses")
	groups.Use(middleware.Auth(jwt))
	{
		groups.GET("", h.List)
		groups.POST("", h.Create)
		groups.GET("/export", h.ExportCSV)
	}
	expenses := r.Group("/expenses")
	expenses.Use(middleware.Auth(jwt))
	{
		expenses.GET("/:id", h.Get)
		expenses.PUT("/:id", h.Update)
		expenses.DELETE("/:id", h.Delete)
	}
}
