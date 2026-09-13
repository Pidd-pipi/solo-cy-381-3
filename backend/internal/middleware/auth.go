package middleware

import (
	"strings"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// Auth 校验 JWT，并把用户 ID、角色、用户名写入上下文。
func Auth(jwt *util.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			util.Fail(c, util.NewAppError(constants.CodeUnauthorized, constants.MsgErrUnauthorized, nil))
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwt.Parse(tokenStr)
		if err != nil {
			util.Fail(c, util.NewAppError(constants.CodeUnauthorized, constants.MsgErrUnauthorized, err))
			c.Abort()
			return
		}
		c.Set(util.CtxUserID, claims.UserID)
		c.Set(util.CtxRole, claims.Role)
		c.Set(util.CtxUsername, claims.Username)
		c.Next()
	}
}
