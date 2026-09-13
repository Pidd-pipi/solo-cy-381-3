package dto

// CreateGroupReq 创建群组请求。
type CreateGroupReq struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"omitempty,max=512"`
}

// UpdateGroupReq 更新群组请求。
type UpdateGroupReq struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"omitempty,max=512"`
}

// InviteMemberReq 邀请成员请求（按用户名或邮箱）。
type InviteMemberReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
}

// GroupResp 群组响应。
type GroupResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     uint   `json:"owner_id"`
	Status      string `json:"status"`
	MemberCount int64  `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

// MemberResp 群组成员响应。
type MemberResp struct {
	ID        uint   `json:"id"`
	GroupID   uint   `json:"group_id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	InvitedBy uint   `json:"invited_by"`
	JoinedAt  string `json:"joined_at"`
}
