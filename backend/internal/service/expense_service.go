package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/aasplit/aasplit/pkg/splitcalc"
	"gorm.io/gorm"
)

// ExpenseService 消费记录业务逻辑。
type ExpenseService struct {
	db          *gorm.DB
	expenseRepo *repository.ExpenseRepository
	memberRepo  *repository.GroupMemberRepository
	groupRepo   *repository.GroupRepository
	userRepo    *repository.UserRepository
	settleRepo  *repository.SettlementRepository
	auditSvc    *AuditService
	logger      *slog.Logger
}

// NewExpenseService 构造消费记录服务。
func NewExpenseService(db *gorm.DB, expenseRepo *repository.ExpenseRepository, memberRepo *repository.GroupMemberRepository, groupRepo *repository.GroupRepository, userRepo *repository.UserRepository, settleRepo *repository.SettlementRepository, auditSvc *AuditService, logger *slog.Logger) *ExpenseService {
	return &ExpenseService{db: db, expenseRepo: expenseRepo, memberRepo: memberRepo, groupRepo: groupRepo, userRepo: userRepo, settleRepo: settleRepo, auditSvc: auditSvc, logger: logger}
}

// Create 创建消费记录并计算分摊明细（单一事务：群组行锁 → 校验/计算 → 写记录 → 写分摊明细 → 失效旧结算建议）。
func (s *ExpenseService) Create(userID uint, req *dto.CreateExpenseReq) (*model.Expense, error) {
	var created *model.Expense
	err := s.db.Transaction(func(tx *gorm.DB) error {
		group, err := s.groupRepo.LockByID(tx, req.GroupID)
		if err != nil {
			return err
		}
		if !group.IsActive() {
			return util.NewAppError(constants.CodeConflict, "群组 group 已归档，无法添加消费记录 expense", nil)
		}
		if err := s.ensureMember(tx, req.GroupID, userID); err != nil {
			return err
		}
		if err := s.ensureMember(tx, req.GroupID, req.PayerID); err != nil {
			return util.NewAppError(constants.CodeBadRequest, fmt.Sprintf("付款人 payer_id=%d 不是群组成员 member", req.PayerID), err)
		}
		shares, err := s.calcShares(tx, req.GroupID, req.Amount, req.SplitType, req.Shares)
		if err != nil {
			return err
		}
		paidAt, err := parsePaidAt(req.PaidAt)
		if err != nil {
			return util.NewAppError(constants.CodeValidationFailed, fmt.Sprintf("消费时间 paid_at 格式无效: %s", req.PaidAt), err)
		}
		expense := &model.Expense{
			GroupID:    req.GroupID,
			Title:      req.Title,
			Amount:     util.Round2(req.Amount),
			Category:   constants.ExpenseCategory(req.Category),
			PayerID:    req.PayerID,
			SplitType:  constants.SplitType(req.SplitType),
			PaidAt:     paidAt,
			ReceiptURL: req.ReceiptURL,
			Status:     constants.ExpenseActive,
			CreatedBy:  userID,
		}
		if err := s.expenseRepo.Create(tx, expense); err != nil {
			return err
		}
		shareModels := make([]model.ExpenseShare, 0, len(shares))
		for _, sh := range shares {
			shareModels = append(shareModels, model.ExpenseShare{
				ExpenseID:   expense.ID,
				UserID:      sh.UserID,
				ShareAmount: sh.ShareAmount,
				Ratio:       sh.Ratio,
				Status:      model.ShareUnsettled,
			})
		}
		if err := s.expenseRepo.CreateShares(tx, shareModels); err != nil {
			return err
		}
		// 金额构成已变化：同一事务内失效该群组的待结算建议，避免返回过期数据。
		if err := s.settleRepo.DeletePendingByGroup(tx, req.GroupID); err != nil {
			return err
		}
		created = expense
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(fmt.Sprintf(constants.LogExpenseCreated, created.ID, created.GroupID, created.Title, created.Amount, created.Category, created.SplitType, created.PayerID))
	s.auditSvc.Record(userID, string(constants.ActionExpenseCreate), "expense", fmt.Sprint(created.ID), "创建消费记录 "+created.Title, "")
	return s.Get(userID, created.ID)
}

// Update 更新消费记录及其分摊明细（单一事务：消费记录行锁 → 群组行锁 → 校验/计算 → 更新记录 → 重建分摊明细 → 失效旧结算建议）。
func (s *ExpenseService) Update(userID, expenseID uint, req *dto.UpdateExpenseReq) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		expense, err := s.expenseRepo.LockByID(tx, expenseID)
		if err != nil {
			return util.Wrap(constants.CodeNotFound, "消费记录 expense 不存在", err)
		}
		group, err := s.groupRepo.LockByID(tx, expense.GroupID)
		if err != nil {
			return util.Wrap(constants.CodeNotFound, "群组 group 不存在", err)
		}
		if !group.IsActive() {
			return util.NewAppError(constants.CodeConflict, "群组 group 已归档，无法修改消费记录 expense", nil)
		}
		if err := s.ensureMember(tx, expense.GroupID, userID); err != nil {
			return err
		}
		if expense.Status != constants.ExpenseActive {
			return util.NewAppError(constants.CodeConflict, "消费记录 expense 已退款，无法修改", nil)
		}
		if err := s.ensureMember(tx, expense.GroupID, req.PayerID); err != nil {
			return util.NewAppError(constants.CodeBadRequest, fmt.Sprintf("付款人 payer_id=%d 不是群组成员 member", req.PayerID), err)
		}
		shares, err := s.calcShares(tx, expense.GroupID, req.Amount, req.SplitType, req.Shares)
		if err != nil {
			return err
		}
		paidAt, err := parsePaidAt(req.PaidAt)
		if err != nil {
			return util.NewAppError(constants.CodeValidationFailed, fmt.Sprintf("消费时间 paid_at 格式无效: %s", req.PaidAt), err)
		}
		expense.Title = req.Title
		expense.Amount = util.Round2(req.Amount)
		expense.Category = constants.ExpenseCategory(req.Category)
		expense.PayerID = req.PayerID
		expense.SplitType = constants.SplitType(req.SplitType)
		expense.PaidAt = paidAt
		expense.ReceiptURL = req.ReceiptURL
		if err := s.expenseRepo.Update(tx, expense); err != nil {
			return err
		}
		if err := s.expenseRepo.DeleteShares(tx, expenseID); err != nil {
			return err
		}
		shareModels := make([]model.ExpenseShare, 0, len(shares))
		for _, sh := range shares {
			shareModels = append(shareModels, model.ExpenseShare{
				ExpenseID:   expenseID,
				UserID:      sh.UserID,
				ShareAmount: sh.ShareAmount,
				Ratio:       sh.Ratio,
				Status:      model.ShareUnsettled,
			})
		}
		if err := s.expenseRepo.CreateShares(tx, shareModels); err != nil {
			return err
		}
		// 金额构成已变化：同一事务内失效该群组的待结算建议，避免返回过期数据。
		if err := s.settleRepo.DeletePendingByGroup(tx, expense.GroupID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.logger.Info(fmt.Sprintf(constants.LogExpenseUpdated, expenseID, req.Title, req.Amount, req.SplitType, userID))
	s.auditSvc.Record(userID, string(constants.ActionExpenseUpdate), "expense", fmt.Sprint(expenseID), "更新消费记录 "+req.Title, "")
	return nil
}

// Delete 将消费记录标记为已退款（单一事务：消费记录行锁 → 群组行锁 → 成员/状态校验 → 状态更新并失效旧结算建议；保留历史）。
// 群组行锁与结算生成互斥：生成建议同样先锁群组行，从而消除“退款删除 pending 后、
// 仍在进行中的生成事务再插入按退款前余额计算的建议”这一并发窗口。
// 分摊明细不删除：统计与结算均通过 expenses.status='active' 过滤，退款记录与其明细作为历史保留。
func (s *ExpenseService) Delete(userID, expenseID uint) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		expense, err := s.expenseRepo.LockByID(tx, expenseID)
		if err != nil {
			return util.Wrap(constants.CodeNotFound, "消费记录 expense 不存在", err)
		}
		if _, err := s.groupRepo.LockByID(tx, expense.GroupID); err != nil {
			return util.Wrap(constants.CodeNotFound, "群组 group 不存在", err)
		}
		if err := s.ensureMember(tx, expense.GroupID, userID); err != nil {
			return err
		}
		if expense.Status != constants.ExpenseActive {
			return util.NewAppError(constants.CodeConflict, "消费记录 expense 已退款，无需重复操作", nil)
		}
		if err := s.expenseRepo.UpdateStatus(tx, expenseID, string(constants.ExpenseRefunded)); err != nil {
			return err
		}
		// 退款改变余额：同一事务内失效该群组的待结算建议，避免返回过期数据。
		if err := s.settleRepo.DeletePendingByGroup(tx, expense.GroupID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.logger.Info(fmt.Sprintf(constants.LogExpenseDeleted, expenseID, 0, userID))
	s.auditSvc.Record(userID, string(constants.ActionExpenseDelete), "expense", fmt.Sprint(expenseID), "退款消费记录", "")
	return nil
}

// Get 查询消费记录详情。
func (s *ExpenseService) Get(userID, expenseID uint) (*model.Expense, error) {
	expense, err := s.expenseRepo.FindByID(nil, expenseID)
	if errors.Is(err, repository.ErrExpenseNotFound) {
		return nil, util.NewAppError(constants.CodeNotFound, "消费记录 expense 不存在", err)
	}
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	ok, err := s.memberRepo.Exists(nil, expense.GroupID, userID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !ok {
		return nil, util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
	}
	return expense, nil
}

// List 分页筛选查询（复用：详情/列表/导出共用同一查询逻辑）。
func (s *ExpenseService) List(userID, groupID uint, query *dto.ExpenseQuery) ([]model.Expense, int64, error) {
	ok, err := s.memberRepo.Exists(nil, groupID, userID)
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !ok {
		return nil, 0, util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
	}
	page, pageSize := query.Page, query.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = util.DefaultPageSize
	}
	params := repository.ExpenseQueryParams{
		GroupID:  groupID,
		Category: query.Category,
		Status:   string(constants.ExpenseActive),
		Page:     page,
		PageSize: pageSize,
	}
	if query.Start != "" {
		if t, err := parsePaidAt(query.Start); err == nil {
			params.Start = t
		}
	}
	if query.End != "" {
		if t, err := parsePaidAt(query.End); err == nil {
			params.End = t
		}
	}
	return s.expenseRepo.List(params)
}

// ExportCSV 导出群组账单明细 CSV（复用 List 的查询逻辑）。
func (s *ExpenseService) ExportCSV(userID, groupID uint, query *dto.ExpenseQuery) ([]byte, string, error) {
	fullQuery := *query
	fullQuery.Page = 1
	fullQuery.PageSize = 10000
	expenses, _, err := s.List(userID, groupID, &fullQuery)
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"ID", "标题", "金额", "类别", "付款人", "分摊方式", "消费时间", "状态"}); err != nil {
		return nil, "", util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	for _, e := range expenses {
		paidAt := e.PaidAt.Format("2006-01-02 15:04:05")
		row := []string{
			strconv.FormatUint(uint64(e.ID), 10),
			e.Title,
			strconv.FormatFloat(e.Amount, 'f', 2, 64),
			util.CategoryText(string(e.Category)),
			e.PayerName(),
			util.SplitTypeText(string(e.SplitType)),
			paidAt,
			util.ExpenseStatusText(string(e.Status)),
		}
		if err := w.Write(row); err != nil {
			return nil, "", util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogExpenseExported, groupID, len(expenses), userID))
	s.auditSvc.Record(userID, string(constants.ActionExpenseExport), "group", fmt.Sprint(groupID), "导出账单明细 CSV", "")
	filename := fmt.Sprintf("expenses_%d_%s.csv", groupID, time.Now().Format("20060102"))
	return buf.Bytes(), filename, nil
}

// calcShares 校验参与人并调用 splitcalc 计算分摊（传入 tx 时，成员读取复用该事务）。
func (s *ExpenseService) calcShares(tx *gorm.DB, groupID uint, amount float64, splitType string, inputs []dto.ShareInput) ([]splitcalc.Share, error) {
	memberIDs, err := s.memberRepo.ListUserIDs(tx, groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	memberSet := make(map[uint]struct{}, len(memberIDs))
	for _, id := range memberIDs {
		memberSet[id] = struct{}{}
	}
	participants := make([]splitcalc.Participant, 0, len(inputs))
	seen := make(map[uint]struct{}, len(inputs))
	for _, in := range inputs {
		if _, dup := seen[in.UserID]; dup {
			return nil, util.NewAppError(constants.CodeValidationFailed, fmt.Sprintf("参与人 user_id=%d 重复", in.UserID), nil)
		}
		seen[in.UserID] = struct{}{}
		if _, ok := memberSet[in.UserID]; !ok {
			return nil, util.NewAppError(constants.CodeBadRequest, fmt.Sprintf("参与人 user_id=%d 不是群组成员 member", in.UserID), nil)
		}
		participants = append(participants, splitcalc.Participant{UserID: in.UserID, Ratio: in.Ratio, Amount: in.Amount})
	}
	shares, err := splitcalc.CalculateShares(amount, splitcalc.SplitType(splitType), participants)
	if errors.Is(err, splitcalc.ErrSumMismatch) {
		return nil, util.NewAppError(constants.CodeExpenseShareMismatch, constants.MsgErrShareMismatch, err)
	}
	if errors.Is(err, splitcalc.ErrInvalidSplit) {
		return nil, util.NewAppError(constants.CodeExpenseInvalidSplit, constants.MsgErrExpenseInvalid, err)
	}
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return shares, nil
}

// ensureMember 校验用户是群组成员。
func (s *ExpenseService) ensureMember(tx *gorm.DB, groupID, userID uint) error {
	var n int64
	if err := tx.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, "active").Count(&n).Error; err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if n == 0 {
		return util.NewAppError(constants.CodeNotGroupMember, fmt.Sprintf("用户 user_id=%d 不是群组 group_id=%d 的成员 member", userID, groupID), nil)
	}
	return nil
}

// parsePaidAt 解析消费时间。
func parsePaidAt(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05Z07:00", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid paid_at format")
}
