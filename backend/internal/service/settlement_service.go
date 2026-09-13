package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/aasplit/aasplit/pkg/splitcalc"
	"gorm.io/gorm"
)

// SettlementService 结算建议业务逻辑。
type SettlementService struct {
	db         *gorm.DB
	settleRepo *repository.SettlementRepository
	shareRepo  *repository.ExpenseShareRepository
	memberRepo *repository.GroupMemberRepository
	groupRepo  *repository.GroupRepository
	userRepo   *repository.UserRepository
	auditSvc   *AuditService
	logger     *slog.Logger
}

// NewSettlementService 构造结算服务。
func NewSettlementService(db *gorm.DB, settleRepo *repository.SettlementRepository, shareRepo *repository.ExpenseShareRepository, memberRepo *repository.GroupMemberRepository, groupRepo *repository.GroupRepository, userRepo *repository.UserRepository, auditSvc *AuditService, logger *slog.Logger) *SettlementService {
	return &SettlementService{db: db, settleRepo: settleRepo, shareRepo: shareRepo, memberRepo: memberRepo, groupRepo: groupRepo, userRepo: userRepo, auditSvc: auditSvc, logger: logger}
}

// Generate 生成智能结算建议（事务 + 群组行锁 + 清除旧建议）。
func (s *SettlementService) Generate(userID, groupID uint) ([]model.Settlement, error) {
	var result []model.Settlement
	err := s.db.Transaction(func(tx *gorm.DB) error {
		group, err := s.groupRepo.LockByID(groupID)
		if err != nil {
			return util.Wrap(constants.CodeNotFound, "群组 group 不存在", err)
		}
		if !group.IsActive() {
			return util.NewAppError(constants.CodeConflict, "群组 group 已归档，无法生成结算建议 settlement", nil)
		}
		memberIDs, err := s.memberRepo.ListUserIDs(groupID)
		if err != nil {
			return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
		}
		if len(memberIDs) < 2 {
			return util.NewAppError(constants.CodeBadRequest, "群组成员 member 不足 2 人，无法生成结算建议 settlement", nil)
		}
		paid, err := s.shareRepo.SumPaidByGroup(groupID)
		if err != nil {
			return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
		}
		owed, err := s.shareRepo.SumOwedByGroup(groupID)
		if err != nil {
			return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
		}
		balances := make([]splitcalc.Balance, 0, len(memberIDs))
		for _, mid := range memberIDs {
			balances = append(balances, splitcalc.Balance{UserID: mid, Amount: util.Round2(paid[mid] - owed[mid])})
		}
		transfers := splitcalc.OptimizeTransfers(balances)
		items := make([]model.Settlement, 0, len(transfers))
		for _, tr := range transfers {
			items = append(items, model.Settlement{
				GroupID:    groupID,
				FromUserID: tr.FromUserID,
				ToUserID:   tr.ToUserID,
				Amount:     tr.Amount,
				Status:     constants.SettlementPending,
			})
		}
		if err := s.settleRepo.DeleteAllByGroup(tx, groupID); err != nil {
			return err
		}
		if err := s.settleRepo.CreateBatch(tx, items); err != nil {
			return err
		}
		result = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 重新查询以加载 FromUser/ToUser 信息
	result, err = s.settleRepo.ListByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if err != nil {
		return nil, err
	}
	s.logger.Info(fmt.Sprintf(constants.LogSettlementGenerated, groupID, len(result), userID))
	s.auditSvc.Record(userID, string(constants.ActionSettlementGenerate), "group", fmt.Sprint(groupID), "生成智能结算建议", "")
	return result, nil
}

// ListByGroup 查询群组结算建议。
func (s *SettlementService) ListByGroup(userID, groupID uint) ([]model.Settlement, error) {
	ok, err := s.memberRepo.Exists(groupID, userID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !ok {
		return nil, util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
	}
	items, err := s.settleRepo.ListByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return items, nil
}

// ListPending 查询我的待结算提醒（from 或 to 涉及当前用户）。
func (s *SettlementService) ListPending(userID uint) ([]model.Settlement, error) {
	items, err := s.settleRepo.ListPendingByUser(userID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return items, nil
}

// Settle 标记结算建议为已结算（事务；仅限群组成员操作）。
func (s *SettlementService) Settle(userID uint, req *dto.SettleReq) (int64, error) {
	var affected int64
	var first *model.Settlement
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range req.SettlementIDs {
			item, err := s.settleRepo.FindByID(id)
			if err != nil {
				return util.Wrap(constants.CodeNotFound, "结算建议 settlement 不存在", err)
			}
			ok, err := s.memberRepo.Exists(item.GroupID, userID)
			if err != nil {
				return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
			}
			if !ok {
				return util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
			}
			if first == nil {
				first = item
			}
		}
		res := tx.Model(&model.Settlement{}).
			Where("id IN ? AND status = ?", req.SettlementIDs, "pending").
			Update("status", "settled")
		if res.Error != nil {
			return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, res.Error)
		}
		affected = res.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	if first != nil {
		s.logger.Info(fmt.Sprintf(constants.LogSettlementSettled, first.ID, first.FromUserID, first.ToUserID, first.Amount, userID))
	}
	ids := make([]string, 0, len(req.SettlementIDs))
	for _, id := range req.SettlementIDs {
		ids = append(ids, fmt.Sprint(id))
	}
	s.auditSvc.Record(userID, string(constants.ActionSettlementSettle), "settlement", strings.Join(ids, ","), "标记结算建议为已结算", "")
	return affected, nil
}

// Balances 计算群组成员净余额（统计页/结算页共用）。
func (s *SettlementService) Balances(userID, groupID uint) ([]dto.GroupBalance, error) {
	ok, err := s.memberRepo.Exists(groupID, userID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !ok {
		return nil, util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
	}
	members, err := s.memberRepo.ListByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	paid, err := s.shareRepo.SumPaidByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	owed, err := s.shareRepo.SumOwedByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	balances := make([]dto.GroupBalance, 0, len(members))
	for _, m := range members {
		if m.User == nil {
			continue
		}
		balances = append(balances, dto.GroupBalance{
			UserID:    m.UserID,
			Username:  m.User.Username,
			Nickname:  m.User.Nickname,
			NetAmount: util.Round2(paid[m.UserID] - owed[m.UserID]),
		})
	}
	return balances, nil
}
