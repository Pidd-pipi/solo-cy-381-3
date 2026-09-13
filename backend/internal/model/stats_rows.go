package model

// CategoryStatRow 类别统计行（repository 扫描结果）。
type CategoryStatRow struct {
	Category string  `gorm:"column:category"`
	Amount   float64 `gorm:"column:amount"`
	Count    int64   `gorm:"column:count"`
}

// MonthlyStatRow 月度统计行（repository 扫描结果）。
type MonthlyStatRow struct {
	Month  string  `gorm:"column:month"`
	Amount float64 `gorm:"column:amount"`
	Count  int64   `gorm:"column:count"`
}
