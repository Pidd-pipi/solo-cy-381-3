package util

import (
	"net/http"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/gin-gonic/gin"
)

// Response 统一响应结构 { code, message, data }。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 统一分页响应结构。
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Created 返回创建成功响应。
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "ok", Data: data})
}

// Fail 返回业务失败响应，错误码由 AppError 决定。
func Fail(c *gin.Context, err error) {
	ae := AsAppError(err)
	c.JSON(ae.Code/100, Response{Code: ae.Code, Message: ae.Message})
}

// ValidationCode 参数校验失败统一错误码。
func ValidationCode(err error) int {
	return constants.CodeValidationFailed
}

// ValidationMessage 参数校验失败文案（含字段名与具体错误）。
func ValidationMessage(err error) string {
	return constants.MsgErrValidation + ": " + err.Error()
}
