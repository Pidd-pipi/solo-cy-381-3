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
	"gorm.io/gorm"
)

// GroupService 分账群组业务逻辑。
type GroupService struct {
	db         *gorm.DB
	groupRepo  *repository.GroupRepository
	memberRepo *repository.GroupMemberRepository
	userRepo   *repository.UserRepository
	auditSvc   *AuditService
	logger     *slog.Logger
}

// NewGroupService 构造群组服务。
func NewGroupService(db *gorm.DB, groupRepo *repository.GroupRepository, memberRepo *repository.GroupMemberRepository, userRepo *repository.UserRepository, auditSvc *AuditService, logger *slog.Logger) *GroupService {
	return &GroupService{db: db, groupRepo: groupRepo, memberRepo: memberRepo, userRepo: userRepo, auditSvc: auditSvc, logger: logger}
}

// Create 创建群组，并把创建者加入为群主（事务）。
func (s *GroupService) Create(userID uint, req *dto.CreateGroupReq) (*model.Group, error) {
	var created *model.Group
	err := s.db.Transaction(func(tx *gorm.DB) error {
		group := &model.Group{
			Name:        req.Name,
			Description: req.Description,
			OwnerID:     userID,
			Status:      constants.GroupActive,
		}
		if err := s.groupRepo.Create(group); err != nil {
			return err
		}
		member := &model.GroupMember{
			GroupID:   group.ID,
			UserID:    userID,
			Role:      model.MemberRoleOwner,
			InvitedBy: userID,
			Status:    constants.GroupActive,
		}
		if err := s.memberRepo.Create(member); err != nil {
			return err
		}
		created = group
		return nil
	})
	if err != nil {
		return nil, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGroupCreated, created.ID, created.Name, userID))
	s.auditSvc.Record(userID, string(constants.ActionGroupCreate), "group", fmt.Sprint(created.ID), "创建群组 "+created.Name, "")
	return created, nil
}

// Update 更新群组（群主/管理员）。
func (s *GroupService) Update(operatorID, groupID uint, req *dto.UpdateGroupReq) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return util.Wrap(constants.CodeNotFound, constants.MsgErrNotFound, err)
	}
	if !s.canManage(operatorID, groupID, group) {
		return util.NewAppError(constants.CodeForbidden, "只有群主或管理员可以修改群组 group 信息", nil)
	}
	group.Name = req.Name
	group.Description = req.Description
	if err := s.groupRepo.Update(group); err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGroupUpdated, group.ID, group.Name, group.Description, operatorID))
	s.auditSvc.Record(operatorID, string(constants.ActionGroupUpdate), "group", fmt.Sprint(groupID), "更新群组 "+group.Name, "")
	return nil
}

// Archive 归档群组（群主/管理员）。
func (s *GroupService) Archive(operatorID, groupID uint) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return util.Wrap(constants.CodeNotFound, constants.MsgErrNotFound, err)
	}
	if !s.canManage(operatorID, groupID, group) {
		return util.NewAppError(constants.CodeForbidden, "只有群主或管理员可以归档群组 group", nil)
	}
	if err := s.groupRepo.UpdateFields(groupID, map[string]interface{}{"status": constants.GroupArchived}); err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogGroupArchived, groupID, operatorID))
	s.auditSvc.Record(operatorID, string(constants.ActionGroupArchive), "group", fmt.Sprint(groupID), "归档群组 "+group.Name, "")
	return nil
}

// Get 查询群组详情（需为成员）。
func (s *GroupService) Get(userID, groupID uint) (*model.Group, int64, error) {
	group, err := s.groupRepo.FindByID(groupID)
	if errors.Is(err, repository.ErrGroupNotFound) {
		return nil, 0, util.NewAppError(constants.CodeNotFound, "群组 group 不存在", err)
	}
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	ok, err := s.memberRepo.Exists(groupID, userID)
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if !ok {
		return nil, 0, util.NewAppError(constants.CodeNotGroupMember, constants.MsgErrNotGroupMember, nil)
	}
	count, err := s.memberRepo.Count(groupID)
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return group, count, nil
}

// ListMy 查询我的群组列表。
func (s *GroupService) ListMy(userID uint, page, pageSize int) ([]model.Group, int64, error) {
	groups, total, err := s.groupRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	return groups, total, nil
}

// InviteMember 邀请成员加入群组（按用户名）。
func (s *GroupService) InviteMember(operatorID, groupID uint, username string) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return util.Wrap(constants.CodeNotFound, constants.MsgErrNotFound, err)
	}
	if !group.IsActive() {
		return util.NewAppError(constants.CodeConflict, "群组 group 已归档，无法邀请成员", nil)
	}
	if !s.canManage(operatorID, groupID, group) {
		return util.NewAppError(constants.CodeForbidden, "只有群主或管理员可以邀请群组成员 member", nil)
	}
	invitee, err := s.userRepo.FindByUsername(username)
	if errors.Is(err, repository.ErrUserNotFound) {
		return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("被邀请用户 %s 不存在", username), err)
	}
	if err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	exists, err := s.memberRepo.Exists(groupID, invitee.ID)
	if err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if exists {
		return util.NewAppError(constants.CodeMemberExists, fmt.Sprintf("用户 %s 已是该群组 group 的成员 member", invitee.Username), nil)
	}
	if err := s.memberRepo.Create(&model.GroupMember{
		GroupID:   groupID,
		UserID:    invitee.ID,
		Role:      model.MemberRoleNormal,
		InvitedBy: operatorID,
		Status:    constants.GroupActive,
	}); err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMemberInvited, groupID, invitee.ID, operatorID))
	s.auditSvc.Record(operatorID, string(constants.ActionMemberInvite), "group_member", fmt.Sprintf("%d:%d", groupID, invitee.ID), "邀请成员 "+invitee.Username+" 加入群组 "+group.Name, "")
	return nil
}

// ListMembers 查询群组成员列表。
func (s *GroupService) ListMembers(userID, groupID uint) ([]model.GroupMember, error) {
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
	return members, nil
}

// RemoveMember 移除群组成员（群主/管理员）。
func (s *GroupService) RemoveMember(operatorID, groupID, userID uint) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return util.Wrap(constants.CodeNotFound, constants.MsgErrNotFound, err)
	}
	if !s.canManage(operatorID, groupID, group) {
		return util.NewAppError(constants.CodeForbidden, "只有群主或管理员可以移除群组成员 member", nil)
	}
	if group.OwnerID == userID {
		return util.NewAppError(constants.CodeConflict, "不能移除群主 owner，请先转移群组", nil)
	}
	member, err := s.memberRepo.Find(groupID, userID)
	if errors.Is(err, repository.ErrMemberNotFound) {
		return util.NewAppError(constants.CodeMemberNotFound, "群组成员 member 不存在", err)
	}
	if err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	if err := s.memberRepo.Remove(groupID, userID); err != nil {
		return util.Wrap(constants.CodeInternalError, constants.MsgErrInternal, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMemberRemoved, groupID, userID, operatorID))
	s.auditSvc.Record(operatorID, string(constants.ActionMemberRemove), "group_member", fmt.Sprintf("%d:%d", groupID, userID), fmt.Sprintf("移除成员 user_id=%d", member.UserID), "")
	return nil
}

// MemberCount 查询群组成员数（供 handler 组装响应复用）。
func (s *GroupService) MemberCount(groupID uint) (int64, error) {
	return s.memberRepo.Count(groupID)
}

// canManage 判断是否群主或管理员。
func (s *GroupService) canManage(userID, groupID uint, group *model.Group) bool {
	if group.OwnerID == userID {
		return true
	}
	u, err := s.userRepo.FindByID(userID)
	if err == nil && u.IsAdmin() {
		return true
	}
	return false
}
