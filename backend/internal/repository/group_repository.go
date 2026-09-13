package repository

import (
	"errors"
	"fmt"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// ErrGroupNotFound 群组不存在哨兵错误。
var ErrGroupNotFound = errors.New("group not found")

// GroupRepository 分账群组数据访问。
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository 构造群组仓储。
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create 创建群组。
func (r *GroupRepository) Create(group *model.Group) error {
	if err := r.db.Create(group).Error; err != nil {
		return fmt.Errorf("create group: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询。
func (r *GroupRepository) FindByID(id uint) (*model.Group, error) {
	var g model.Group
	if err := r.db.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("find group by id: %w", err)
	}
	return &g, nil
}

// Update 更新群组。
func (r *GroupRepository) Update(group *model.Group) error {
	if err := r.db.Save(group).Error; err != nil {
		return fmt.Errorf("update group: %w", err)
	}
	return nil
}

// UpdateFields 按字段图更新群组。
func (r *GroupRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	if err := r.db.Model(&model.Group{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		return fmt.Errorf("update group fields: %w", err)
	}
	return nil
}

// LockByID 行级锁查询群组（并发写场景 SELECT ... FOR UPDATE）。
func (r *GroupRepository) LockByID(id uint) (*model.Group, error) {
	var g model.Group
	if err := r.db.Clauses(lockClause).First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("lock group by id: %w", err)
	}
	return &g, nil
}

// ListByUser 查询用户参与的群组（含成员数）。
func (r *GroupRepository) ListByUser(userID uint, page, pageSize int) ([]model.Group, int64, error) {
	var groups []model.Group
	var total int64
	sub := r.db.Model(&model.GroupMember{}).Select("group_id").Where("user_id = ? AND status = ?", userID, "active")
	if err := r.db.Model(&model.Group{}).Where("id IN (?)", sub).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count user groups: %w", err)
	}
	if err := r.db.Where("id IN (?)", sub).Order("updated_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("list user groups: %w", err)
	}
	return groups, total, nil
}

// CountMembers 统计群组成员数。
func (r *GroupRepository) CountMembers(groupID uint) (int64, error) {
	var n int64
	if err := r.db.Model(&model.GroupMember{}).Where("group_id = ? AND status = ?", groupID, "active").Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count group members: %w", err)
	}
	return n, nil
}
