package handler

import (
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// StatsHandler 数据统计 HTTP 处理器。
type StatsHandler struct {
	statsSvc *service.StatsService
}

// NewStatsHandler 构造统计处理器。
func NewStatsHandler(statsSvc *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

// GetGroupStats 查询群组统计。
func (h *StatsHandler) GetGroupStats(c *gin.Context) {
	groupID := parseIDParam(c)
	stats, err := h.statsSvc.GetGroupStats(util.GetUserID(c), groupID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, stats)
}
