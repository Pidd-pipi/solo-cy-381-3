package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// lockClause 行级排他锁，用于并发写场景 SELECT ... FOR UPDATE。
var lockClause = clause.Locking{Strength: "UPDATE"}

// txOrDB 返回事务句柄 tx；tx 为空时回退到仓储默认连接 db。
// 所有可在事务内复用的读写方法都通过它保证使用同一事务边界。
func txOrDB(tx, db *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return db
}
