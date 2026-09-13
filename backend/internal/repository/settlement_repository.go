package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// ErrSettlementNotFound 结算建议不存在哨兵错误。
var ErrSettlementNotFound = errors.New("settlement not found")

// SettlementRepository 结算建议数据访问。
type SettlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository 构造结算建议仓储。
func NewSettlementRepository(db *gorm.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

// CreateBatch 批量创建结算建议（事务中调用）。
func (r *SettlementRepository) CreateBatch(tx *gorm.DB, items []model.Settlement) error {
	if len(items) == 0 {
		return nil
	}
	if err := tx.Create(&items).Error; err != nil {
		return fmt.Errorf("create settlements: %w", err)
	}
	return nil
}

// ListByGroup 查询群组结算建议（含双方用户信息）。
func (r *SettlementRepository) ListByGroup(groupID uint) ([]model.Settlement, error) {
	var items []model.Settlement
	if err := r.db.Preload("FromUser").Preload("ToUser").
		Where("group_id = ?", groupID).Order("status ASC, id ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list settlements by group: %w", err)
	}
	return items, nil
}

// ListPendingByUser 查询用户待结算的结算建议。
func (r *SettlementRepository) ListPendingByUser(userID uint) ([]model.Settlement, error) {
	var items []model.Settlement
	if err := r.db.Preload("FromUser").Preload("ToUser").
		Where("(from_user_id = ? OR to_user_id = ?) AND status = ?", userID, userID, "pending").
		Order("id DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list pending settlements: %w", err)
	}
	return items, nil
}

// FindByID 按 ID 查询。
func (r *SettlementRepository) FindByID(id uint) (*model.Settlement, error) {
	var s model.Settlement
	if err := r.db.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettlementNotFound
		}
		return nil, fmt.Errorf("find settlement by id: %w", err)
	}
	return &s, nil
}

// MarkSettled 将指定 ID 列表的结算建议标记为已结算。
func (r *SettlementRepository) MarkSettled(tx *gorm.DB, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	if err := tx.Model(&model.Settlement{}).
		Where("id IN ? AND status = ?", ids, "pending").
		Updates(map[string]interface{}{"status": "settled", "settled_at": now}).Error; err != nil {
		return fmt.Errorf("mark settlements settled: %w", err)
	}
	return nil
}

// DeleteAllByGroup 清空群组结算建议（重新生成前调用）。
func (r *SettlementRepository) DeleteAllByGroup(tx *gorm.DB, groupID uint) error {
	if err := tx.Where("group_id = ?", groupID).Delete(&model.Settlement{}).Error; err != nil {
		return fmt.Errorf("delete settlements by group: %w", err)
	}
	return nil
}
