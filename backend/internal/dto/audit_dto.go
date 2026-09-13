package dto

// AuditQuery 审计日志查询。
type AuditQuery struct {
	Page     int    `form:"page" binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Action   string `form:"action" binding:"omitempty"`
	UserID   uint   `form:"user_id" binding:"omitempty,gte=0"`
}

// AuditLogResp 审计日志响应。
type AuditLogResp struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Detail       string `json:"detail"`
	IP           string `json:"ip"`
	RequestID    string `json:"request_id"`
	CreatedAt    string `json:"created_at"`
}
