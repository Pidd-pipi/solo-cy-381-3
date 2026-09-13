// Package router 按实体注册路由。
package router

import (
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户与认证路由。
func RegisterUserRoutes(r *gin.RouterGroup, h *handler.UserHandler, jwt *util.JWTManager) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
	users := r.Group("/users")
	users.Use(middleware.Auth(jwt))
	{
		users.GET("/me", h.GetMe)
		users.PUT("/me", h.UpdateProfile)
		users.PUT("/me/password", h.ChangePassword)
		users.GET("", middleware.RequireRoles("admin"), h.ListUsers)
		users.PUT("/:id/role", middleware.RequireRoles("admin"), h.ChangeRole)
	}
}
