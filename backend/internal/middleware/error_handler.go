package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// ErrorHandler 统一异常恢复：捕获 panic、请求体解析错误，统一转为标准 JSON 响应。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Error(fmt.Sprintf(constants.LogRecoveredPanic, util.GetRequestID(c), r, string(stack)))
				util.Fail(c, util.NewAppError(constants.CodeInternalError, constants.MsgErrInternal, nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// NoRoute 处理未匹配路由。
func NoRoute(c *gin.Context) {
	util.Fail(c, util.NewAppError(constants.CodeNotFound, "接口不存在: "+c.Request.Method+" "+c.Request.URL.Path, nil))
}

// NoMethod 处理方法不允许。
func NoMethod(c *gin.Context) {
	util.Fail(c, util.NewAppError(http.StatusMethodNotAllowed*100, "请求方法不允许: "+c.Request.Method, nil))
}
