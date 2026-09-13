package repository

import (
	"fmt"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// ExpenseShareRepository 分摊明细数据访问。
type ExpenseShareRepository struct {
	db *gorm.DB
}

// NewExpenseShareRepository 构造分摊明细仓储。
func NewExpenseShareRepository(db *gorm.DB) *ExpenseShareRepository {
	return &ExpenseShareRepository{db: db}
}

// SumPaidByGroup 统计群组内各成员付款总额（付款人维度）。
func (r *ExpenseShareRepository) SumPaidByGroup(groupID uint) (map[uint]float64, error) {
	type row struct {
		UserID uint
		Total  float64
	}
	var rows []row
	if err := r.db.Model(&model.Expense{}).
		Select("payer_id AS user_id, SUM(amount) AS total").
		Where("group_id = ? AND status = ?", groupID, "active").
		Group("payer_id").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("sum paid by group: %w", err)
	}
	result := make(map[uint]float64, len(rows))
	for _, rw := range rows {
		result[rw.UserID] = rw.Total
	}
	return result, nil
}

// SumOwedByGroup 统计群组内各成员应付总额（参与人维度，仅有效消费）。
func (r *ExpenseShareRepository) SumOwedByGroup(groupID uint) (map[uint]float64, error) {
	type row struct {
		UserID uint
		Total  float64
	}
	var rows []row
	if err := r.db.Table("expense_shares AS s").
		Joins("JOIN expenses AS e ON e.id = s.expense_id").
		Select("s.user_id AS user_id, SUM(s.share_amount) AS total").
		Where("e.group_id = ? AND e.status = ? AND s.status = ?", groupID, "active", "unsettled").
		Group("s.user_id").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("sum owed by group: %w", err)
	}
	result := make(map[uint]float64, len(rows))
	for _, rw := range rows {
		result[rw.UserID] = rw.Total
	}
	return result, nil
}

// MarkSettledByExpenses 将某批消费的分摊明细标记为已结算。
func (r *ExpenseShareRepository) MarkSettledByExpenses(tx *gorm.DB, expenseIDs []uint) error {
	if len(expenseIDs) == 0 {
		return nil
	}
	if err := tx.Model(&model.ExpenseShare{}).
		Where("expense_id IN ?", expenseIDs).
		Update("status", "settled").Error; err != nil {
		return fmt.Errorf("mark shares settled: %w", err)
	}
	return nil
}

// CountUnsettledByUser 统计用户待结算的分摊明细条数。
func (r *ExpenseShareRepository) CountUnsettledByUser(userID uint) (int64, error) {
	var n int64
	if err := r.db.Table("expense_shares AS s").
		Joins("JOIN expenses AS e ON e.id = s.expense_id").
		Where("s.user_id = ? AND s.status = ? AND e.status = ?", userID, "unsettled", "active").
		Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count unsettled shares: %w", err)
	}
	return n, nil
}
