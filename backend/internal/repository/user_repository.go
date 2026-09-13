// Package repository 封装数据访问，仓储层定义哨兵错误并向上透传。
package repository

import (
	"errors"
	"fmt"

	"github.com/aasplit/aasplit/internal/model"
	"gorm.io/gorm"
)

// ErrUserNotFound 用户不存在哨兵错误。
var ErrUserNotFound = errors.New("user not found")

// UserRepository 用户数据访问。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户。
func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// FindByUsername 按用户名查询。
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &u, nil
}

// FindByEmail 按邮箱查询。
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

// FindByID 按 ID 查询。
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

// Update 更新用户字段。
func (r *UserRepository) Update(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// UpdateFields 按字段图更新用户。
func (r *UserRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	if err := r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		return fmt.Errorf("update user fields: %w", err)
	}
	return nil
}

// List 分页查询用户（管理员）。
func (r *UserRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	query := r.db.Model(&model.User{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

// CountByUsername 统计同名用户数量。
func (r *UserRepository) CountByUsername(username string) (int64, error) {
	var n int64
	if err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count username: %w", err)
	}
	return n, nil
}

// CountByEmail 统计同邮箱用户数量。
func (r *UserRepository) CountByEmail(email string) (int64, error) {
	var n int64
	if err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count email: %w", err)
	}
	return n, nil
}
