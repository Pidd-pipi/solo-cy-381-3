// Package service 承载业务逻辑，依赖注入仓储，多步写操作放入事务。
package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/util"
)

// UserService 用户业务逻辑。
type UserService struct {
	userRepo *repository.UserRepository
	jwt      *util.JWTManager
	logger   *slog.Logger
}

// NewUserService 构造用户服务。
func NewUserService(userRepo *repository.UserRepository, jwt *util.JWTManager, logger *slog.Logger) *UserService {
	return &UserService{userRepo: userRepo, jwt: jwt, logger: logger}
}

// Register 注册新用户。
func (s *UserService) Register(req *dto.RegisterReq) (*model.User, error) {
	if n, err := s.userRepo.CountByUsername(req.Username); err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	} else if n > 0 {
		return nil, util.NewAppError(constants.CodeUsernameExists, constants.MsgErrUsernameExists, nil)
	}
	if req.Email != "" {
		if n, err := s.userRepo.CountByEmail(req.Email); err != nil {
			return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
		} else if n > 0 {
			return nil, util.NewAppError(constants.CodeEmailExists, constants.MsgErrEmailExists, nil)
		}
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, fmt.Errorf("hash password for user %s: %w", req.Username, err))
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		Role:         constants.RoleUser,
		Status:       model.UserStatusActive,
	}
	if req.Email != "" {
		email := req.Email
		user.Email = &email
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserRegistered, user.ID, user.Username, user.Role))
	return user, nil
}

// Login 登录并签发 JWT。
func (s *UserService) Login(req *dto.LoginReq) (string, *model.User, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if errors.Is(err, repository.ErrUserNotFound) {
		s.logger.Warn(fmt.Sprintf(constants.LogLoginFailed, req.Username, "user not found"))
		return "", nil, util.NewAppError(constants.CodeWrongPassword, constants.MsgErrWrongPassword, nil)
	}
	if err != nil {
		return "", nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !util.CheckPassword(user.PasswordHash, req.Password) {
		s.logger.Warn(fmt.Sprintf(constants.LogLoginFailed, req.Username, "wrong password"))
		return "", nil, util.NewAppError(constants.CodeWrongPassword, constants.MsgErrWrongPassword, nil)
	}
	if !user.IsActive() {
		return "", nil, util.NewAppError(constants.CodeUserDisabled, constants.MsgErrUserDisabled, nil)
	}
	token, err := s.jwt.Generate(user.ID, user.Username, string(user.Role))
	if err != nil {
		return "", nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, fmt.Errorf("generate token for user %s: %w", user.Username, err))
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserLoggedIn, user.ID, user.Username))
	return token, user, nil
}

// GetByID 查询用户。
func (s *UserService) GetByID(userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, util.NewAppError(constants.CodeNotFound, constants.MsgErrNotFound, err)
	}
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return user, nil
}

// UpdateProfile 更新昵称/邮箱/头像。
func (s *UserService) UpdateProfile(userID uint, req *dto.UpdateProfileReq) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	oldEmail := ""
	if user.Email != nil {
		oldEmail = *user.Email
	}
	if req.Email != "" && req.Email != oldEmail {
		if n, err2 := s.userRepo.CountByEmail(req.Email); err2 != nil {
			return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err2)
		} else if n > 0 {
			return nil, util.NewAppError(constants.CodeEmailExists, constants.MsgErrEmailExists, nil)
		}
	}
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	if req.Email != "" {
		email := req.Email
		user.Email = &email
	} else {
		user.Email = nil
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserUpdated, user.ID, user.Nickname, user.Avatar))
	return user, nil
}

// ChangePassword 修改密码。
func (s *UserService) ChangePassword(userID uint, req *dto.ChangePasswordReq) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil
	}
	if !util.CheckPassword(user.PasswordHash, req.OldPassword) {
		return util.NewAppError(constants.CodeWrongPassword, constants.MsgErrWrongPassword, nil)
	}
	hash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if err := s.userRepo.UpdateFields(userID, map[string]interface{}{"password_hash": hash}); err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return nil
}

// ChangeRole 管理员修改用户角色。
func (s *UserService) ChangeRole(operatorID, targetID uint, role string) error {
	if !constants.IsValidUserRole(role) {
		return util.NewAppError(constants.CodeBadRequest, "角色参数无效，user 角色仅支持 user/admin", nil)
	}
	if err := s.userRepo.UpdateFields(targetID, map[string]interface{}{"role": role}); err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserRoleChanged, targetID, role, operatorID))
	return nil
}

// List 分页查询用户（管理员）。
func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return users, total, nil
}
