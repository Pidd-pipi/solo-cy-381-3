package util

import (
	"github.com/gin-gonic/gin"
)

const (
	// CtxUserID 当前登录用户 ID 上下文键。
	CtxUserID = "ctx_user_id"
	// CtxRole 当前登录用户角色上下文键。
	CtxRole = "ctx_role"
	// CtxUsername 当前登录用户名上下文键。
	CtxUsername = "ctx_username"
	// CtxRequestID 请求 ID 上下文键。
	CtxRequestID = "ctx_request_id"
	// CtxIP 客户端 IP 上下文键。
	CtxIP = "ctx_ip"
)

// GetUserID 从上下文取当前用户 ID。
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok2 := v.(uint); ok2 {
			return id
		}
	}
	return 0
}

// GetRole 从上下文取当前用户角色。
func GetRole(c *gin.Context) string {
	if v, ok := c.Get(CtxRole); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// GetUsername 从上下文取当前用户名。
func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(CtxUsername); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// GetRequestID 从上下文取请求 ID。
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(CtxRequestID); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// GetClientIP 从上下文取客户端 IP。
func GetClientIP(c *gin.Context) string {
	if v, ok := c.Get(CtxIP); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return c.ClientIP()
}
