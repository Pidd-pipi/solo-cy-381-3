package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseIDParam 解析 :id 路径参数。
func parseIDParam(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}

// parseIDParamName 解析指定名称的路径参数。
func parseIDParamName(c *gin.Context, name string) uint {
	id, _ := strconv.ParseUint(c.Param(name), 10, 64)
	return uint(id)
}
