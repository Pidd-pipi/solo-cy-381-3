// Package util 提供日志、JWT、密码、响应、错误与格式化等通用能力。
package util

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger 创建 slog 结构化日志器，级别由环境变量 LOG_LEVEL 控制。
func NewLogger() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
