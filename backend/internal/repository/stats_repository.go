package repository

import (
	"fmt"
	"time"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// StatsRepository 数据统计访问。
type StatsRepository struct {
	db *gorm.DB
}

// NewStatsRepository 构造统计仓储。
func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// SumByCategory 按类别统计金额与笔数。
func (r *StatsRepository) SumByCategory(groupID uint) ([]model.CategoryStatRow, error) {
	var rows []model.CategoryStatRow
	if err := r.db.Model(&model.Expense{}).
		Select("category, SUM(amount) AS amount, COUNT(*) AS count").
		Where("group_id = ? AND status = ?", groupID, "active").
		Group("category").Order("amount DESC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("sum by category: %w", err)
	}
	return rows, nil
}

// SumByMonth 按月统计金额与笔数（近 12 个月）。
func (r *StatsRepository) SumByMonth(groupID uint) ([]model.MonthlyStatRow, error) {
	var rows []model.MonthlyStatRow
	start := time.Now().AddDate(0, -11, 0)
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
	if err := r.db.Model(&model.Expense{}).
		Select("TO_CHAR(paid_at, 'YYYY-MM') AS month, SUM(amount) AS amount, COUNT(*) AS count").
		Where("group_id = ? AND status = ? AND paid_at >= ?", groupID, "active", start).
		Group("month").Order("month ASC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("sum by month: %w", err)
	}
	return rows, nil
}

// TotalByGroup 统计群组消费总额与笔数。
func (r *StatsRepository) TotalByGroup(groupID uint) (float64, int64, error) {
	type row struct {
		TotalAmount float64
		TotalCount  int64
	}
	var rr row
	if err := r.db.Model(&model.Expense{}).
		Select("COALESCE(SUM(amount),0) AS total_amount, COUNT(*) AS total_count").
		Where("group_id = ? AND status = ?", groupID, "active").Scan(&rr).Error; err != nil {
		return 0, 0, fmt.Errorf("total by group: %w", err)
	}
	return rr.TotalAmount, rr.TotalCount, nil
}
