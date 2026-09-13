package model

// PayerName 返回付款人昵称（关联未加载时回退为空字符串）。
func (e *Expense) PayerName() string {
	if e.Payer != nil {
		return e.Payer.Nickname
	}
	return ""
}
