package constants

// messages.go 同时承载接口返回文案、日志文案、错误提示文案（屎山耦合点 3）。
const (
	MsgOK                  = "ok"
	MsgCreated             = "创建成功"
	MsgUpdated             = "更新成功"
	MsgDeleted             = "删除成功"
	MsgLoginSuccess        = "登录成功"
	MsgRegisterSuccess     = "注册成功"
	MsgGroupCreated        = "群组创建成功"
	MsgMemberInvited       = "成员邀请成功"
	MsgExpenseCreated      = "消费记录创建成功"
	MsgSettlementGenerated = "结算建议生成成功"
	MsgSettlementSettled   = "结算完成"
	MsgExportStarted       = "导出成功"

	MsgErrBind           = "请求体解析失败，请检查字段格式"
	MsgErrValidation     = "参数校验失败，请检查必填字段与取值范围"
	MsgErrUnauthorized   = "登录状态已失效，请重新登录"
	MsgErrForbidden      = "当前角色无权执行该操作"
	MsgErrNotFound       = "请求的资源不存在"
	MsgErrInternal       = "服务内部错误，请稍后重试"
	MsgErrUsernameExists = "用户名已存在，请更换用户名"
	MsgErrEmailExists    = "邮箱已存在，请更换邮箱"
	MsgErrWrongPassword  = "密码错误，请重新输入"
	MsgErrUserDisabled   = "账号已被禁用，请联系管理员"
	MsgErrNotGroupMember = "您不是该群组成员，无法操作"
	MsgErrExpenseInvalid = "分摊参数无效，请检查参与人与分摊配置"
	MsgErrShareMismatch  = "分摊金额合计与消费总额不一致"
	MsgErrRateLimited    = "请求过于频繁，请稍后再试"
)
