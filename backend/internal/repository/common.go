package repository

import "gorm.io/gorm/clause"

// lockClause 行级排他锁，用于并发写场景 SELECT ... FOR UPDATE。
var lockClause = clause.Locking{Strength: "UPDATE"}
