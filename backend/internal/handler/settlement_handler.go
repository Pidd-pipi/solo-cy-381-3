package handler

import (
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// SettlementHandler 结算建议 HTTP 处理器。
type SettlementHandler struct {
	settleSvc *service.SettlementService
}

// NewSettlementHandler 构造结算处理器。
func NewSettlementHandler(settleSvc *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{settleSvc: settleSvc}
}

// Generate 生成智能结算建议。
func (h *SettlementHandler) Generate(c *gin.Context) {
	groupID := parseIDParam(c)
	items, err := h.settleSvc.Generate(util.GetUserID(c), groupID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]*dto.SettlementResp, 0, len(items))
	for i := range items {
		list = append(list, toSettlementResp(&items[i]))
	}
	util.OK(c, gin.H{"list": list, "total": len(list)})
}

// ListByGroup 查询群组结算建议。
func (h *SettlementHandler) ListByGroup(c *gin.Context) {
	groupID := parseIDParam(c)
	items, err := h.settleSvc.ListByGroup(util.GetUserID(c), groupID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]*dto.SettlementResp, 0, len(items))
	for i := range items {
		list = append(list, toSettlementResp(&items[i]))
	}
	util.OK(c, gin.H{"list": list, "total": len(list)})
}

// ListPending 我的待结算提醒。
func (h *SettlementHandler) ListPending(c *gin.Context) {
	items, err := h.settleSvc.ListPending(util.GetUserID(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]*dto.SettlementResp, 0, len(items))
	for i := range items {
		list = append(list, toSettlementResp(&items[i]))
	}
	util.OK(c, gin.H{"list": list, "total": len(list)})
}

// Settle 标记结算完成。
func (h *SettlementHandler) Settle(c *gin.Context) {
	var req dto.SettleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	affected, err := h.settleSvc.Settle(util.GetUserID(c), &req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "结算完成", "affected": affected})
}

// Balances 群组成员净余额。
func (h *SettlementHandler) Balances(c *gin.Context) {
	groupID := parseIDParam(c)
	balances, err := h.settleSvc.Balances(util.GetUserID(c), groupID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"list": balances, "total": len(balances)})
}

// toSettlementResp 结算模型转响应。
func toSettlementResp(s *model.Settlement) *dto.SettlementResp {
	fromName, toName := "", ""
	if s.FromUser != nil {
		fromName = s.FromUser.Nickname
	}
	if s.ToUser != nil {
		toName = s.ToUser.Nickname
	}
	resp := &dto.SettlementResp{
		ID:         s.ID,
		GroupID:    s.GroupID,
		FromUserID: s.FromUserID,
		FromName:   fromName,
		ToUserID:   s.ToUserID,
		ToName:     toName,
		Amount:     util.Round2(s.Amount),
		Status:     string(s.Status),
		CreatedAt:  util.FormatDateTime(s.CreatedAt),
	}
	if s.SettledAt != nil {
		resp.SettledAt = util.FormatDateTime(*s.SettledAt)
	}
	return resp
}
