package service

import (
	"log/slog"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
)

// StatsService 数据统计业务逻辑。
type StatsService struct {
	statsRepo  *repository.StatsRepository
	shareRepo  *repository.ExpenseShareRepository
	memberRepo *repository.GroupMemberRepository
	logger     *slog.Logger
}

// NewStatsService 构造统计服务。
func NewStatsService(statsRepo *repository.StatsRepository, shareRepo *repository.ExpenseShareRepository, memberRepo *repository.GroupMemberRepository, logger *slog.Logger) *StatsService {
	return &StatsService{statsRepo: statsRepo, shareRepo: shareRepo, memberRepo: memberRepo, logger: logger}
}

// GetGroupStats 获取群组统计（类别占比、月度趋势、成员排行）。
func (s *StatsService) GetGroupStats(userID, groupID uint) (*dto.StatsResp, error) {
	ok, err := s.memberRepo.Exists(nil, groupID, userID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !ok {
		return nil, util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
	}
	categoryRows, err := s.statsRepo.SumByCategory(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	monthRows, err := s.statsRepo.SumByMonth(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	totalAmount, totalCount, err := s.statsRepo.TotalByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	members, err := s.memberRepo.ListByGroup(groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	paid, err := s.shareRepo.SumPaidByGroup(nil, groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	owed, err := s.shareRepo.SumOwedByGroup(nil, groupID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	ranks := make([]dto.MemberRank, 0, len(members))
	for _, m := range members {
		if m.User == nil {
			continue
		}
		ranks = append(ranks, dto.MemberRank{
			UserID:   m.UserID,
			Username: m.User.Username,
			Nickname: m.User.Nickname,
			Paid:     util.Round2(paid[m.UserID]),
			Owed:     util.Round2(owed[m.UserID]),
			Net:      util.Round2(paid[m.UserID] - owed[m.UserID]),
		})
	}
	categoryStats := make([]dto.CategoryStat, 0, len(categoryRows))
	for _, r := range categoryRows {
		categoryStats = append(categoryStats, dto.CategoryStat{Category: r.Category, Amount: util.Round2(r.Amount), Count: r.Count})
	}
	monthlyStats := make([]dto.MonthlyStat, 0, len(monthRows))
	for _, r := range monthRows {
		monthlyStats = append(monthlyStats, dto.MonthlyStat{Month: r.Month, Amount: util.Round2(r.Amount), Count: r.Count})
	}
	return &dto.StatsResp{
		CategoryStats: categoryStats,
		MonthlyStats:  monthlyStats,
		MemberRanks:   ranks,
		TotalExpense:  util.Round2(totalAmount),
		TotalCount:    totalCount,
	}, nil
}
