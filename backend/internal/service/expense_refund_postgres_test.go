package service

import (
	"os"
	"testing"
	"time"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/model"
	"github.com/aasplit/aasplit/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// 本文件提供针对 PostgreSQL 的确定性并发交错集成测试。
//
// 背景：SQLite 以整库唯一写锁把所有写事务完全串行化，无法复现 PostgreSQL
// READ COMMITTED 下的交错——结算生成事务先读到退款前余额，退款事务随后提交并
// 删除 pending，生成事务最后再插入按旧余额计算的建议。修复手段是退款与生成都
// 在事务内对同一群组行执行 SELECT ... FOR UPDATE，使二者在该行上互斥串行。
//
// 运行方式：提供测试库 DSN 后执行
//
//	TEST_POSTGRES_DSN='host=/tmp user=postgres dbname=postgres sslmode=disable' \
//	  go test ./internal/service/ -run TestPostgresRefundGenerateRace -count=1 -v
//
// 未设置环境变量时自动跳过，不影响常规（无外部依赖的）测试。

// openPostgresTestDB 在设置 TEST_POSTGRES_DSN 时连接并迁移测试库，否则跳过当前测试。
func openPostgresTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set TEST_POSTGRES_DSN to run PostgreSQL concurrency integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(10)
	if err := migrateAll(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() {
		for _, tbl := range []string{"audit_logs", "settlements", "expense_shares", "expenses", "group_members", "groups", "users"} {
			_ = db.Exec("TRUNCATE TABLE " + tbl + " RESTART IDENTITY CASCADE").Error
		}
		_ = sqlDB.Close()
	})
	return db
}

// TestPostgresLockClauseEmitsForUpdate 方言验证：退款/生成所用的行锁查询在
// PostgreSQL 下都生成 SELECT ... FOR UPDATE（ToSQL 只构建语句，不执行）。
func TestPostgresLockClauseEmitsForUpdate(t *testing.T) {
	db := openPostgresTestDB(t)
	for _, stmt := range []struct {
		name string
		sql  string
	}{
		{"group", db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Model(&model.Group{}).Where("id = ?", 1).
				Clauses(clause.Locking{Strength: "UPDATE"}).Find(&[]model.Group{})
		})},
		{"expense", db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Model(&model.Expense{}).Where("id = ?", 1).
				Clauses(clause.Locking{Strength: "UPDATE"}).Find(&[]model.Expense{})
		})},
	} {
		if !containsSubstr(stmt.sql, "FOR UPDATE") {
			t.Fatalf("%s lock SQL missing FOR UPDATE: %s", stmt.name, stmt.sql)
		}
	}
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestPostgresRefundGenerateRace 确定性交错验证（退款侧调用真实 ExpenseService.Delete）：
//   - 生成事务先持有群组行锁；真实退款事务必须被该行锁阻塞；
//   - 生成事务按退款前余额写入旧建议后提交；
//   - 退款随后继续：更新状态并删除 pending，旧建议不得残留；
//   - 重新生成读取最新消费数据，结果为空。
//
// 若退款实现回归（不再锁群组行），退款会在生成插入前完成删除并提交，
// 生成随后插入的旧建议将残留，最终 pending 计数断言失败。
func TestPostgresRefundGenerateRace(t *testing.T) {
	db := openPostgresTestDB(t)

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	memberRepo := repository.NewGroupMemberRepository(db)
	settleRepo := repository.NewSettlementRepository(db)
	expenseSvc := newExpenseServiceFromDB(db)
	settleSvc := buildSettlementService(db)

	u1 := &model.User{Username: "pg_alice", PasswordHash: "x", Nickname: "Alice", Role: constants.RoleUser, Status: "active"}
	u2 := &model.User{Username: "pg_bob", PasswordHash: "x", Nickname: "Bob", Role: constants.RoleUser, Status: "active"}
	u3 := &model.User{Username: "pg_carol", PasswordHash: "x", Nickname: "Carol", Role: constants.RoleUser, Status: "active"}
	for _, u := range []*model.User{u1, u2, u3} {
		if err := userRepo.Create(u); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	g := &model.Group{Name: "pg-group", OwnerID: u1.ID, Status: constants.GroupActive}
	if err := groupRepo.Create(g); err != nil {
		t.Fatalf("create group: %v", err)
	}
	for idx, uid := range []uint{u1.ID, u2.ID, u3.ID} {
		role := model.MemberRoleNormal
		if idx == 0 {
			role = model.MemberRoleOwner
		}
		if err := memberRepo.Create(&model.GroupMember{GroupID: g.ID, UserID: uid, Role: role, Status: constants.GroupActive}); err != nil {
			t.Fatalf("create member: %v", err)
		}
	}
	paidAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	expense := &model.Expense{
		GroupID: g.ID, Title: "火锅", Amount: 300, Category: "dining", PayerID: u1.ID,
		SplitType: "equal", PaidAt: paidAt, Status: constants.ExpenseActive, CreatedBy: u1.ID,
	}
	if err := db.Create(expense).Error; err != nil {
		t.Fatalf("create expense: %v", err)
	}
	shares := []model.ExpenseShare{
		{ExpenseID: expense.ID, UserID: u1.ID, ShareAmount: 100, Ratio: 0.3333, Status: model.ShareUnsettled},
		{ExpenseID: expense.ID, UserID: u2.ID, ShareAmount: 100, Ratio: 0.3333, Status: model.ShareUnsettled},
		{ExpenseID: expense.ID, UserID: u3.ID, ShareAmount: 100, Ratio: 0.3333, Status: model.ShareUnsettled},
	}
	if err := db.Create(&shares).Error; err != nil {
		t.Fatalf("create shares: %v", err)
	}

	// “生成事务”的临界区：独立连接上开启事务并持有群组行锁。
	txG := db.Begin()
	if err := txG.Error; err != nil {
		t.Fatalf("begin generate tx: %v", err)
	}
	if _, err := groupRepo.LockByID(txG, g.ID); err != nil {
		t.Fatalf("generate locks group: %v", err)
	}

	// 真实退款事务（生产代码）：锁消费行后会尝试锁群组行，必须被阻塞。
	refundDone := make(chan error, 1)
	go func() {
		refundDone <- expenseSvc.Delete(u1.ID, expense.ID)
	}()
	select {
	case err := <-refundDone:
		t.Fatalf("refund completed while generate held the group row lock — missing FOR UPDATE: %v", err)
	case <-time.After(300 * time.Millisecond):
		// 预期：退款在等待群组行锁。
	}

	// 生成事务按退款前余额写入旧建议（2 笔合计 200），然后提交释放群组行锁。
	if err := settleRepo.CreateBatch(txG, []model.Settlement{
		{GroupID: g.ID, FromUserID: u2.ID, ToUserID: u1.ID, Amount: 100, Status: constants.SettlementPending},
		{GroupID: g.ID, FromUserID: u3.ID, ToUserID: u1.ID, Amount: 100, Status: constants.SettlementPending},
	}); err != nil {
		t.Fatalf("generate inserts stale suggestions: %v", err)
	}
	if err := txG.Commit().Error; err != nil {
		t.Fatalf("commit generate: %v", err)
	}

	// 群组行锁释放后，真实退款应在超时内完成，并删除刚写入的旧建议。
	select {
	case err := <-refundDone:
		if err != nil {
			t.Fatalf("refund after generate commit: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("refund did not complete after generate committed")
	}

	var pending int64
	if err := db.Model(&model.Settlement{}).Where("group_id = ? AND status = ?", g.ID, "pending").Count(&pending).Error; err != nil {
		t.Fatalf("count pending: %v", err)
	}
	if pending != 0 {
		t.Fatalf("stale pending settlements = %d after refund, want 0", pending)
	}
	var status string
	if err := db.Model(&model.Expense{}).Where("id = ?", expense.ID).Pluck("status", &status).Error; err != nil {
		t.Fatalf("read expense status: %v", err)
	}
	if status != string(constants.ExpenseRefunded) {
		t.Fatalf("expense status = %s, want refunded", status)
	}

	// 重新生成读取最新消费数据：唯一消费已退款，结果应为空。
	items, err := settleSvc.Generate(u1.ID, g.ID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("regenerated transfers = %d, want 0 (latest data after refund)", len(items))
	}
}
