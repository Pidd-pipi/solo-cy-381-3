package handler

import (
	"net/http"
	"strconv"

	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
)

// ExpenseHandler 消费记录 HTTP 处理器。
type ExpenseHandler struct {
	expenseSvc *service.ExpenseService
}

// NewExpenseHandler 构造消费记录处理器。
func NewExpenseHandler(expenseSvc *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseSvc: expenseSvc}
}

// Create 创建消费记录。
func (h *ExpenseHandler) Create(c *gin.Context) {
	var req dto.CreateExpenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	if req.GroupID == 0 {
		req.GroupID = parseIDParam(c)
	}
	expense, err := h.expenseSvc.Create(util.GetUserID(c), &req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.Created(c, toExpenseResp(expense))
}

// Update 更新消费记录。
func (h *ExpenseHandler) Update(c *gin.Context) {
	var req dto.UpdateExpenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	expenseID := parseIDParam(c)
	if err := h.expenseSvc.Update(util.GetUserID(c), expenseID, &req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "消费记录更新成功"})
}

// Delete 退款消费记录。
func (h *ExpenseHandler) Delete(c *gin.Context) {
	expenseID := parseIDParam(c)
	if err := h.expenseSvc.Delete(util.GetUserID(c), expenseID); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"message": "消费记录已退款"})
}

// Get 查询消费记录详情。
func (h *ExpenseHandler) Get(c *gin.Context) {
	expenseID := parseIDParam(c)
	expense, err := h.expenseSvc.Get(util.GetUserID(c), expenseID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, toExpenseResp(expense))
}

// List 分页筛选查询。
func (h *ExpenseHandler) List(c *gin.Context) {
	groupID := parseIDParam(c)
	var query dto.ExpenseQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	expenses, total, err := h.expenseSvc.List(util.GetUserID(c), groupID, &query)
	if err != nil {
		util.Fail(c, err)
		return
	}
	list := make([]*dto.ExpenseResp, 0, len(expenses))
	for i := range expenses {
		list = append(list, toExpenseResp(&expenses[i]))
	}
	page, pageSize := util.ParsePagination(c)
	util.OK(c, util.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// ExportCSV 导出账单明细 CSV。
func (h *ExpenseHandler) ExportCSV(c *gin.Context) {
	groupID := parseIDParam(c)
	var query dto.ExpenseQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		util.Fail(c, util.Wrap(util.ValidationCode(err), util.ValidationMessage(err), err))
		return
	}
	data, filename, err := h.expenseSvc.ExportCSV(util.GetUserID(c), groupID, &query)
	if err != nil {
		util.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// toExpenseResp 消费记录模型转响应。
func toExpenseResp(e *model.Expense) *dto.ExpenseResp {
	resp := &dto.ExpenseResp{
		ID:         e.ID,
		GroupID:    e.GroupID,
		Title:      e.Title,
		Amount:     util.Round2(e.Amount),
		Category:   string(e.Category),
		PayerID:    e.PayerID,
		PayerName:  e.PayerName(),
		SplitType:  string(e.SplitType),
		PaidAt:     util.FormatDateTime(e.PaidAt),
		ReceiptURL: e.ReceiptURL,
		Status:     string(e.Status),
		CreatedBy:  e.CreatedBy,
		CreatedAt:  util.FormatDateTime(e.CreatedAt),
		Shares:     make([]dto.ExpenseShareResp, 0, len(e.Shares)),
	}
	for _, s := range e.Shares {
		username, nickname := "", ""
		if s.User != nil {
			username = s.User.Username
			nickname = s.User.Nickname
		}
		resp.Shares = append(resp.Shares, dto.ExpenseShareResp{
			UserID:      s.UserID,
			Username:    username,
			Nickname:    nickname,
			ShareAmount: util.Round2(s.ShareAmount),
			Ratio:       util.Round2(s.Ratio),
			Status:      string(s.Status),
		})
	}
	return resp
}

var _ = strconv.Itoa
