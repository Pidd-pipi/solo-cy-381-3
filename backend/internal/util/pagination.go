package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// ParsePagination 从 query 解析统一分页参数 page / page_size。
func ParsePagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = DefaultPageSize
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.Query("page_size")); err == nil && v > 0 {
		pageSize = v
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

// Offset 计算分页偏移量。
func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}
