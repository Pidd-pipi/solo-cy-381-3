// Package dto 定义各实体的请求/响应数据结构与校验规则。
package dto

// IDReq 通用 ID 路径参数。
type IDReq struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

// PageQuery 统一分页查询参数。
type PageQuery struct {
	Page     int `form:"page" binding:"omitempty,gte=1"`
	PageSize int `form:"page_size" binding:"omitempty,gte=1,lte=100"`
}
