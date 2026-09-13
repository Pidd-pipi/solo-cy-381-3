package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// ErrExpenseNotFound 消费记录不存在哨兵错误。
var ErrExpenseNotFound = errors.New("expense not found")

// ExpenseQueryParams 消费记录查询参数。
type ExpenseQueryParams struct {
	GroupID  uint
	Category string
	Start    time.Time
	End      time.Time
	Status   string
	Page     int
	PageSize int
}

// ExpenseRepository 消费记录数据访问。
type ExpenseRepository struct {
	db *gorm.DB
}

// NewExpenseRepository 构造消费记录仓储。
func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create 创建消费记录（事务中调用）。
func (r *ExpenseRepository) Create(tx *gorm.DB, expense *model.Expense) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Create(expense).Error; err != nil {
		return fmt.Errorf("create expense: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询（含付款人信息）。
func (r *ExpenseRepository) FindByID(id uint) (*model.Expense, error) {
	var e model.Expense
	if err := r.db.Preload("Payer").Preload("Shares.User").First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExpenseNotFound
		}
		return nil, fmt.Errorf("find expense by id: %w", err)
	}
	return &e, nil
}

// List 分页筛选查询（含付款人与分摊明细）。
func (r *ExpenseRepository) List(p ExpenseQueryParams) ([]model.Expense, int64, error) {
	var expenses []model.Expense
	var total int64
	query := r.db.Model(&model.Expense{}).Where("group_id = ?", p.GroupID)
	if p.Status != "" {
		query = query.Where("status = ?", p.Status)
	}
	if p.Category != "" {
		query = query.Where("category = ?", p.Category)
	}
	if !p.Start.IsZero() {
		query = query.Where("paid_at >= ?", p.Start)
	}
	if !p.End.IsZero() {
		query = query.Where("paid_at <= ?", p.End)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count expenses: %w", err)
	}
	if err := query.Preload("Payer").Preload("Shares.User").
		Order("paid_at DESC").Offset((p.Page - 1) * p.PageSize).Limit(p.PageSize).Find(&expenses).Error; err != nil {
		return nil, 0, fmt.Errorf("list expenses: %w", err)
	}
	return expenses, total, nil
}

// Update 更新消费记录。
func (r *ExpenseRepository) Update(tx *gorm.DB, expense *model.Expense) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Save(expense).Error; err != nil {
		return fmt.Errorf("update expense: %w", err)
	}
	return nil
}

// UpdateStatus 更新消费状态。
func (r *ExpenseRepository) UpdateStatus(tx *gorm.DB, id uint, status string) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Model(&model.Expense{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("update expense status: %w", err)
	}
	return nil
}

// DeleteShares 删除某消费记录的全部分摊明细（事务中调用）。
func (r *ExpenseRepository) DeleteShares(tx *gorm.DB, expenseID uint) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Where("expense_id = ?", expenseID).Delete(&model.ExpenseShare{}).Error; err != nil {
		return fmt.Errorf("delete expense shares: %w", err)
	}
	return nil
}

// CreateShares 批量创建分摊明细（事务中调用）。
func (r *ExpenseRepository) CreateShares(tx *gorm.DB, shares []model.ExpenseShare) error {
	if len(shares) == 0 {
		return nil
	}
	if tx == nil {
		tx = r.db
	}
	if err := tx.Create(&shares).Error; err != nil {
		return fmt.Errorf("create expense shares: %w", err)
	}
	return nil
}
