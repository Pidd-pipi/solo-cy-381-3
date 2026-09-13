// Package middleware 提供认证、RBAC、请求追踪、错误处理、限流、审计等横切能力。
package middleware

import (
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 为每个请求注入 X-Request-ID（缺失时生成），并写入上下文。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(util.CtxRequestID, rid)
		c.Set(util.CtxIP, c.ClientIP())
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
