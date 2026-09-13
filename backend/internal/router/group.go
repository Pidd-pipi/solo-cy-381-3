package router

import (
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RegisterGroupRoutes 注册群组路由。
func RegisterGroupRoutes(r *gin.RouterGroup, h *handler.GroupHandler, jwt *util.JWTManager) {
	groups := r.Group("/groups")
	groups.Use(middleware.Auth(jwt))
	{
		groups.GET("", h.ListMy)
		groups.POST("", h.Create)
		groups.GET("/:id", h.Get)
		groups.PUT("/:id", h.Update)
		groups.POST("/:id/archive", h.Archive)
		groups.POST("/:id/members", h.InviteMember)
		groups.GET("/:id/members", h.ListMembers)
		groups.DELETE("/:id/members/:userId", h.RemoveMember)
	}
}
