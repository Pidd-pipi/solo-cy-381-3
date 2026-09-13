package util

import (
	"errors"
	"fmt"

	"github.com/aasplit/aasplit/internal/constants"
)

// AppError 业务错误：包含错误码与面向用户的 message（由 handler/middleware 转为标准 JSON）。
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s cause=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误。
func NewAppError(code int, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Err: cause}
}

// Wrap 包装底层错误并透传，保留错误链。
func Wrap(code int, message string, cause error) error {
	if cause == nil {
		return NewAppError(code, message, nil)
	}
	return NewAppError(code, message, cause)
}

// AsAppError 将任意 error 转为 AppError；非 AppError 一律转为内部错误。
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return NewAppError(constants.CodeInternalError, constants.MsgErrInternal, err)
}

// IsNotFound 判断错误链中是否包含未找到哨兵错误。
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// ErrNotFound 仓储层哨兵错误：资源不存在。
var ErrNotFound = errors.New("resource not found")
