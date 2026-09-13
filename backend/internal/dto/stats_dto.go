package dto

// CategoryStat 类别占比统计项。
type CategoryStat struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int64   `json:"count"`
}

// MonthlyStat 月度趋势统计项。
type MonthlyStat struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

// MemberRank 成员消费排行项。
type MemberRank struct {
	UserID   uint    `json:"user_id"`
	Username string  `json:"username"`
	Nickname string  `json:"nickname"`
	Paid     float64 `json:"paid"`
	Owed     float64 `json:"owed"`
	Net      float64 `json:"net"`
}

// StatsResp 数据统计响应。
type StatsResp struct {
	CategoryStats []CategoryStat `json:"category_stats"`
	MonthlyStats  []MonthlyStat  `json:"monthly_stats"`
	MemberRanks   []MemberRank   `json:"member_ranks"`
	TotalExpense  float64        `json:"total_expense"`
	TotalCount    int64          `json:"total_count"`
}
