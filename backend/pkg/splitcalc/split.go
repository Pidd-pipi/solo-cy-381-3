// Package splitcalc 提供 AA 分摊计算与智能结算路径优化算法（无业务依赖，可复用）。
package splitcalc

import (
	"errors"
	"sort"
)

// SplitType 分摊方式。
type SplitType string

const (
	SplitEqual  SplitType = "equal"  // 均摊
	SplitRatio  SplitType = "ratio"  // 按比例
	SplitAmount SplitType = "amount" // 按金额
)

// Participant 分摊参与人。
type Participant struct {
	UserID uint    `json:"user_id"`
	Ratio  float64 `json:"ratio,omitempty"`  // ratio 模式下占比
	Amount float64 `json:"amount,omitempty"` // amount 模式下应付金额
}

// Share 单人的分摊结果。
type Share struct {
	UserID      uint    `json:"user_id"`
	ShareAmount float64 `json:"share_amount"`
	Ratio       float64 `json:"ratio"`
}

// ErrInvalidSplit 分摊参数无效。
var ErrInvalidSplit = errors.New("invalid split parameters")

// ErrSumMismatch 分摊金额合计与消费总额不匹配。
var ErrSumMismatch = errors.New("share sum mismatch")

// round2 四舍五入保留两位小数。
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// CalculateShares 根据分摊方式计算每人应付金额；保证合计等于总额（最后一人吸收舍入误差）。
func CalculateShares(total float64, splitType SplitType, participants []Participant) ([]Share, error) {
	if total <= 0 {
		return nil, ErrInvalidSplit
	}
	if len(participants) == 0 {
		return nil, ErrInvalidSplit
	}
	shares := make([]Share, 0, len(participants))
	switch splitType {
	case SplitEqual:
		base := round2(total / float64(len(participants)))
		allocated := 0.0
		for i, p := range participants {
			amount := base
			if i == len(participants)-1 {
				amount = round2(total - allocated)
			}
			shares = append(shares, Share{UserID: p.UserID, ShareAmount: amount, Ratio: round2(1 / float64(len(participants)))})
			allocated = round2(allocated + amount)
		}
	case SplitRatio:
		sumRatio := 0.0
		for _, p := range participants {
			if p.Ratio < 0 {
				return nil, ErrInvalidSplit
			}
			sumRatio += p.Ratio
		}
		if sumRatio <= 0 {
			return nil, ErrInvalidSplit
		}
		allocated := 0.0
		for i, p := range participants {
			amount := round2(total * p.Ratio / sumRatio)
			if i == len(participants)-1 {
				amount = round2(total - allocated)
			}
			shares = append(shares, Share{UserID: p.UserID, ShareAmount: amount, Ratio: round2(p.Ratio / sumRatio)})
			allocated = round2(allocated + amount)
		}
	case SplitAmount:
		sumAmount := 0.0
		for _, p := range participants {
			if p.Amount < 0 {
				return nil, ErrInvalidSplit
			}
			sumAmount += p.Amount
		}
		if round2(sumAmount) != round2(total) {
			return nil, ErrSumMismatch
		}
		for _, p := range participants {
			ratio := 0.0
			if total > 0 {
				ratio = round2(p.Amount / total)
			}
			shares = append(shares, Share{UserID: p.UserID, ShareAmount: round2(p.Amount), Ratio: ratio})
		}
	default:
		return nil, ErrInvalidSplit
	}
	return shares, nil
}

// Balance 成员净余额（正数应收，负数应付）。
type Balance struct {
	UserID uint
	Amount float64
}

// Transfer 一条结算转账。
type Transfer struct {
	FromUserID uint
	ToUserID   uint
	Amount     float64
}

// OptimizeTransfers 基于净余额贪心匹配最大债权人与最大债务人，最小化转账次数。
func OptimizeTransfers(balances []Balance) []Transfer {
	creditors := make([]Balance, 0, len(balances))
	debtors := make([]Balance, 0, len(balances))
	for _, b := range balances {
		if round2(b.Amount) > 0 {
			creditors = append(creditors, b)
		} else if round2(b.Amount) < 0 {
			debtors = append(debtors, Balance{UserID: b.UserID, Amount: -b.Amount})
		}
	}
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].Amount > creditors[j].Amount })
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].Amount > debtors[j].Amount })

	transfers := make([]Transfer, 0, len(balances))
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		pay := round2(debtors[i].Amount)
		recv := round2(creditors[j].Amount)
		amount := pay
		if recv < pay {
			amount = recv
		}
		amount = round2(amount)
		if amount > 0 {
			transfers = append(transfers, Transfer{FromUserID: debtors[i].UserID, ToUserID: creditors[j].UserID, Amount: amount})
		}
		debtors[i].Amount = round2(debtors[i].Amount - amount)
		creditors[j].Amount = round2(creditors[j].Amount - amount)
		if round2(debtors[i].Amount) <= 0 {
			i++
		}
		if round2(creditors[j].Amount) <= 0 {
			j++
		}
	}
	return transfers
}
