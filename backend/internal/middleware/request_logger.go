package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// RequestLogger 结构化请求日志，包含 request_id/method/path/status/latency_ms。
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rid := util.GetRequestID(c)
		logger.Info(fmt.Sprintf(constants.LogRequestIn, rid, c.Request.Method, c.Request.URL.Path))
		c.Next()
		latency := time.Since(start).Milliseconds()
		logger.Info(fmt.Sprintf(constants.LogRequestDone, rid, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency))
	}
}
