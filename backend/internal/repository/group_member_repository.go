package repository

import (
	"errors"
	"fmt"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// ErrMemberNotFound 群组成员不存在哨兵错误。
var ErrMemberNotFound = errors.New("group member not found")

// GroupMemberRepository 群组成员数据访问。
type GroupMemberRepository struct {
	db *gorm.DB
}

// NewGroupMemberRepository 构造群组成员仓储。
func NewGroupMemberRepository(db *gorm.DB) *GroupMemberRepository {
	return &GroupMemberRepository{db: db}
}

// Create 添加成员。
func (r *GroupMemberRepository) Create(member *model.GroupMember) error {
	if err := r.db.Create(member).Error; err != nil {
		return fmt.Errorf("create group member: %w", err)
	}
	return nil
}

// Find 查询指定群组成员关系。
func (r *GroupMemberRepository) Find(groupID, userID uint) (*model.GroupMember, error) {
	var m model.GroupMember
	if err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("find group member: %w", err)
	}
	return &m, nil
}

// Exists 判断成员关系是否存在（传 tx 时复用该事务）。
func (r *GroupMemberRepository) Exists(tx *gorm.DB, groupID, userID uint) (bool, error) {
	var n int64
	if err := txOrDB(tx, r.db).Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, "active").Count(&n).Error; err != nil {
		return false, fmt.Errorf("check group member exists: %w", err)
	}
	return n > 0, nil
}

// ListByGroup 查询群组全部有效成员（含用户信息）。
func (r *GroupMemberRepository) ListByGroup(groupID uint) ([]model.GroupMember, error) {
	var members []model.GroupMember
	if err := r.db.Preload("User").Where("group_id = ? AND status = ?", groupID, "active").
		Order("id ASC").Find(&members).Error; err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	return members, nil
}

// ListUserIDs 查询群组有效成员 ID 列表（传 tx 时复用该事务）。
func (r *GroupMemberRepository) ListUserIDs(tx *gorm.DB, groupID uint) ([]uint, error) {
	var ids []uint
	if err := txOrDB(tx, r.db).Model(&model.GroupMember{}).Where("group_id = ? AND status = ?", groupID, "active").
		Pluck("user_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("list group member ids: %w", err)
	}
	return ids, nil
}

// Remove 软删除成员关系。
func (r *GroupMemberRepository) Remove(groupID, userID uint) error {
	if err := r.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("status", "archived").Error; err != nil {
		return fmt.Errorf("remove group member: %w", err)
	}
	return nil
}

// Count 统计群组有效成员数。
func (r *GroupMemberRepository) Count(groupID uint) (int64, error) {
	var n int64
	if err := r.db.Model(&model.GroupMember{}).Where("group_id = ? AND status = ?", groupID, "active").Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count group members: %w", err)
	}
	return n, nil
}
