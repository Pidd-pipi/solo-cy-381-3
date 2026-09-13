package util

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aasplit/aasplit/internal/constants"
)

// formatters.go 同时包含日期、状态文本、类型文本等格式化逻辑（屎山耦合点 3/4/5）。

// Round2 四舍五入保留两位小数。
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// FormatMoney 格式化金额为两位小数字符串。
func FormatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

// FormatDateTime 格式化时间为 yyyy-MM-dd HH:mm:ss。
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化时间为 yyyy-MM-dd。
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// CategoryText 将消费类别枚举转为中文文案。
func CategoryText(c string) string {
	switch constants.ExpenseCategory(c) {
	case constants.CategoryDining:
		return "餐饮"
	case constants.CategoryTransport:
		return "交通"
	case constants.CategoryLodging:
		return "住宿"
	case constants.CategoryEntertain:
		return "娱乐"
	default:
		return "其他"
	}
}

// SplitTypeText 将分摊方式枚举转为中文文案。
func SplitTypeText(s string) string {
	switch constants.SplitType(s) {
	case constants.SplitEqual:
		return "均摊"
	case constants.SplitRatio:
		return "按比例"
	case constants.SplitAmount:
		return "按金额"
	default:
		return "未知"
	}
}

// GroupStatusText 将群组状态枚举转为中文文案。
func GroupStatusText(s string) string {
	switch constants.GroupStatus(s) {
	case constants.GroupActive:
		return "进行中"
	case constants.GroupArchived:
		return "已归档"
	default:
		return "未知"
	}
}

// ExpenseStatusText 将消费状态枚举转为中文文案。
func ExpenseStatusText(s string) string {
	switch constants.ExpenseStatus(s) {
	case constants.ExpenseActive:
		return "有效"
	case constants.ExpenseRefunded:
		return "已退款"
	default:
		return "未知"
	}
}

// SettlementStatusText 将结算状态枚举转为中文文案。
func SettlementStatusText(s string) string {
	switch constants.SettlementStatus(s) {
	case constants.SettlementPending:
		return "待结算"
	case constants.SettlementSettled:
		return "已结算"
	default:
		return "未知"
	}
}

// RoleText 将角色枚举转为中文文案。
func RoleText(r string) string {
	if constants.UserRole(r) == constants.RoleAdmin {
		return "管理员"
	}
	return "普通用户"
}

// TitleSnake 将英文标题转为 snake_case，用于生成资源类型。
func TitleSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + 32)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
