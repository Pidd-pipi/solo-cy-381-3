package handler

import (
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// GroupHandler 分账群组 HTTP 处理器。
type GroupHandler struct {
	groupSvc *service.GroupService
}

// NewGroupHandler 构造群组处理器。
func NewGroupHandler(groupSvc *service.GroupService) *GroupHandler {
	return &GroupHandler{groupSvc: groupSvc}
}

// Create 创建群组。
func (h *GroupHandler) Create(c *gin.Context) {
	var req dto.CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	group, err := h.groupSvc.Create(util.GetUserID(c), &req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.Created(c, toGroupResp(group, 1))
}

// Update 更新群组。
func (h *GroupHandler) Update(c *gin.Context) {
	var req dto.UpdateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	groupID := parseIDParam(c)
	if err := h.groupSvc.Update(util.GetUserID(c), groupID, &req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "群组更新成功"})
}

// Archive 归档群组。
func (h *GroupHandler) Archive(c *gin.Context) {
	groupID := parseIDParam(c)
	if err := h.groupSvc.Archive(util.GetUserID(c), groupID); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "群组已归档"})
}

// Get 查询群组详情。
func (h *GroupHandler) Get(c *gin.Context) {
	groupID := parseIDParam(c)
	group, count, err := h.groupSvc.Get(util.GetUserID(c), groupID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, toGroupResp(group, count))
}

// ListMy 我的群组列表。
func (h *GroupHandler) ListMy(c *gin.Context) {
	page, pageSize := util.ParsePagination(c)
	groups, total, err := h.groupSvc.ListMy(util.GetUserID(c), page, pageSize)
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]*dto.GroupResp, 0, len(groups))
	for i := range groups {
		count, err2 := h.groupSvc.MemberCount(groups[i].ID)
		if err2 != nil {
			count = 0
		}
		list = append(list, toGroupResp(&groups[i], count))
	}
	util.OK(c, util.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// InviteMember 邀请成员。
func (h *GroupHandler) InviteMember(c *gin.Context) {
	var req dto.InviteMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	groupID := parseIDParam(c)
	if err := h.groupSvc.InviteMember(util.GetUserID(c), groupID, req.Username); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "成员邀请成功"})
}

// ListMembers 成员列表。
func (h *GroupHandler) ListMembers(c *gin.Context) {
	groupID := parseIDParam(c)
	members, err := h.groupSvc.ListMembers(util.GetUserID(c), groupID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]dto.MemberResp, 0, len(members))
	for _, m := range members {
		list = append(list, toMemberResp(&m))
	}
	util.OK(c, gin.H{"list": list, "total": len(list)})
}

// RemoveMember 移除成员。
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	groupID := parseIDParam(c)
	userID := parseIDParamName(c, "userId")
	if err := h.groupSvc.RemoveMember(util.GetUserID(c), groupID, userID); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "成员已移除"})
}

// toGroupResp 群组模型转响应。
func toGroupResp(g *model.Group, memberCount int64) *dto.GroupResp {
	return &dto.GroupResp{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		OwnerID:     g.OwnerID,
		Status:      string(g.Status),
		MemberCount: memberCount,
		CreatedAt:   util.FormatDateTime(g.CreatedAt),
	}
}

// toMemberResp 成员模型转响应。
func toMemberResp(m *model.GroupMember) dto.MemberResp {
	username, nickname, avatar := "", "", ""
	if m.User != nil {
		username = m.User.Username
		nickname = m.User.Nickname
		avatar = m.User.Avatar
	}
	return dto.MemberResp{
		ID:        m.ID,
		GroupID:   m.GroupID,
		UserID:    m.UserID,
		Username:  username,
		Nickname:  nickname,
		Avatar:    avatar,
		Role:      string(m.Role),
		InvitedBy: m.InvitedBy,
		JoinedAt:  util.FormatDateTime(m.CreatedAt),
	}
}
