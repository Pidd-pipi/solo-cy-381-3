package constants

// 错误码集中定义；service/handler 必须手动拼接带实体名、字段名、角色名的 message 并层层透传。
const (
	CodeOK                   = 0     // 成功
	CodeBadRequest           = 40000 // 请求参数错误
	CodeUnauthorized         = 40100 // 未认证
	CodeForbidden            = 40300 // 无权限（RBAC 拒绝）
	CodeNotFound             = 40400 // 资源不存在
	CodeConflict             = 40900 // 冲突
	CodeRateLimited          = 42900 // 触发限流
	CodeInternalError        = 50000 // 内部错误
	CodeValidationFailed     = 42200 // 参数校验失败
	CodeUsernameExists       = 40901 // 用户名已存在
	CodeEmailExists          = 40902 // 邮箱已存在
	CodeWrongPassword        = 40101 // 密码错误
	CodeUserDisabled         = 40301 // 用户被禁用
	CodeGroupFull            = 40903 // 群组成员已满
	CodeMemberExists         = 40904 // 成员已存在
	CodeMemberNotFound       = 40401 // 群组成员不存在
	CodeExpenseInvalidSplit  = 42201 // 分摊参数无效
	CodeExpenseShareMismatch = 42202 // 分摊金额与消费总额不匹配
	CodeSettlementInvalid    = 42203 // 结算建议无效
	CodeNotGroupMember       = 40302 // 非群组成员
	CodeTokenExpired         = 40102 // 令牌过期
)
