package middleware

import (
	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RequireRoles 校验当前用户角色；未满足返回 403（RBAC 权限控制）。
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := util.GetRole(c)
		if _, ok := allowed[role]; !ok {
			util.Fail(c, util.NewAppError(constants.CodeForbidden, constants.MsgErrForbidden, nil))
			c.Abort()
			return
		}
		c.Next()
	}
}
