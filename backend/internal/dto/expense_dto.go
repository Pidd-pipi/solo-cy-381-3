package dto

// ShareInput 分摊参与人输入。
type ShareInput struct {
	UserID uint    `json:"user_id" binding:"required,gt=0"`
	Ratio  float64 `json:"ratio" binding:"omitempty,gte=0"`
	Amount float64 `json:"amount" binding:"omitempty,gte=0"`
}

// CreateExpenseReq 创建消费记录请求（group_id 可省略，由路径参数补充）。
type CreateExpenseReq struct {
	GroupID    uint         `json:"group_id" binding:"omitempty,gt=0"`
	Title      string       `json:"title" binding:"required,min=1,max=128"`
	Amount     float64      `json:"amount" binding:"required,gt=0"`
	Category   string       `json:"category" binding:"required,oneof=dining transport lodging entertain other"`
	PayerID    uint         `json:"payer_id" binding:"required,gt=0"`
	SplitType  string       `json:"split_type" binding:"required,oneof=equal ratio amount"`
	PaidAt     string       `json:"paid_at" binding:"required"`
	ReceiptURL string       `json:"receipt_url" binding:"omitempty,max=512"`
	Shares     []ShareInput `json:"shares" binding:"required,min=1"`
}

// UpdateExpenseReq 更新消费记录请求。
type UpdateExpenseReq struct {
	Title      string       `json:"title" binding:"required,min=1,max=128"`
	Amount     float64      `json:"amount" binding:"required,gt=0"`
	Category   string       `json:"category" binding:"required,oneof=dining transport lodging entertain other"`
	PayerID    uint         `json:"payer_id" binding:"required,gt=0"`
	SplitType  string       `json:"split_type" binding:"required,oneof=equal ratio amount"`
	PaidAt     string       `json:"paid_at" binding:"required"`
	ReceiptURL string       `json:"receipt_url" binding:"omitempty,max=512"`
	Shares     []ShareInput `json:"shares" binding:"required,min=1"`
}

// ExpenseQuery 消费记录筛选查询。
type ExpenseQuery struct {
	Page     int    `form:"page" binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Category string `form:"category" binding:"omitempty,oneof=dining transport lodging entertain other"`
	Start    string `form:"start" binding:"omitempty"`
	End      string `form:"end" binding:"omitempty"`
}

// ExpenseShareResp 分摊明细响应。
type ExpenseShareResp struct {
	UserID      uint    `json:"user_id"`
	Username    string  `json:"username"`
	Nickname    string  `json:"nickname"`
	ShareAmount float64 `json:"share_amount"`
	Ratio       float64 `json:"ratio"`
	Status      string  `json:"status"`
}

// ExpenseResp 消费记录响应。
type ExpenseResp struct {
	ID         uint               `json:"id"`
	GroupID    uint               `json:"group_id"`
	Title      string             `json:"title"`
	Amount     float64            `json:"amount"`
	Category   string             `json:"category"`
	PayerID    uint               `json:"payer_id"`
	PayerName  string             `json:"payer_name"`
	SplitType  string             `json:"split_type"`
	PaidAt     string             `json:"paid_at"`
	ReceiptURL string             `json:"receipt_url"`
	Status     string             `json:"status"`
	CreatedBy  uint               `json:"created_by"`
	CreatedAt  string             `json:"created_at"`
	Shares     []ExpenseShareResp `json:"shares"`
}
