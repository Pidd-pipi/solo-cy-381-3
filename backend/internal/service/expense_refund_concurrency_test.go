package service

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/aasplit/aasplit/internal/dto"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openSQLiteFile 打开指向同一文件库的句柄；immediate=true 时事务一开始即取写锁。
func openSQLiteFile(t *testing.T, path string, immediate bool, maxConns int) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)", path)
	if immediate {
		dsn += "&_txlock=immediate"
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite %s: %v", path, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if maxConns > 0 {
		sqlDB.SetMaxOpenConns(maxConns)
	}
	return db
}

// newConcurrentFixture 构造并发测试夹具：
// 建表与初始数据走 deferred 句柄（兼容现有服务的事务用法）；
// 返回的业务服务绑定在 immediate 句柄上——事务开始即持有全库唯一写锁，
// 多个 goroutine 借 busy_timeout 串行提交，近似 PostgreSQL 行锁的串行化（SQLite 下更严格）。
func newConcurrentFixture(t *testing.T) (*gorm.DB, *ExpenseService, *SettlementService, uint, uint, uint, uint) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "concurrent.db")
	setupDB := openSQLiteFile(t, path, false, 0)
	if err := migrateAll(setupDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	_, _, _, groupID, aliceID, bobID, carolID := newExpenseServiceFixtureDB(t, setupDB)

	db := openSQLiteFile(t, path, true, 8)
	sqlDB, _ := db.DB()
	// 不复用可能残留写锁状态的空闲连接；事务期间各 goroutine 仍有独立连接。
	sqlDB.SetMaxIdleConns(0)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db, newExpenseServiceFromDB(db), buildSettlementService(db), groupID, aliceID, bobID, carolID
}

// newExpenseServiceFromDB 在指定句柄上装配消费服务（复用夹具的构造方式）。
func newExpenseServiceFromDB(db *gorm.DB) *ExpenseService {
	return NewExpenseService(
		db,
		repository.NewExpenseRepository(db),
		repository.NewGroupMemberRepository(db),
		repository.NewGroupRepository(db),
		repository.NewUserRepository(db),
		repository.NewSettlementRepository(db),
		NewAuditService(repository.NewAuditRepository(db), newTestLogger()),
		newTestLogger(),
	)
}

// 本文件验证“退款与结算生成并发”时的状态一致性：
// 退款与生成都会在事务内对群组行加锁（PostgreSQL 上为 SELECT ... FOR UPDATE），
// 二者因此串行化。退款提交后，数据库中残留的待结算建议必须只反映最新消费数据，
// 绝不能留下按退款前余额计算的建议。
//
// 测试库为 SQLite：以立即事务 + busy_timeout 近似 PostgreSQL 的写串行化；
// FOR UPDATE 在 SQLite 上是空操作，真正的并发保证由生产代码的群组行锁在 PostgreSQL 上提供。

// pendingTransferTotal 汇总群组待结算建议的转账金额。
func pendingTransferTotal(t *testing.T, settleSvc *SettlementService, groupID, userID uint) float64 {
	t.Helper()
	items, err := settleSvc.ListByGroup(userID, groupID)
	if err != nil {
		t.Fatalf("list by group: %v", err)
	}
	total := 0.0
	for _, it := range items {
		if it.Status == "pending" {
			total += it.Amount
		}
	}
	return total
}

// TestRefundConcurrentGenerate 退款与多次生成并发：最终待结算建议只能是“无消费”或“按最新余额”，不能出现退款前金额。
func TestRefundConcurrentGenerate(t *testing.T) {
	for iter := 0; iter < 20; iter++ {
		db, expenseSvc, settleSvc, groupID, aliceID, bobID, carolID := newConcurrentFixture(t)
		_ = db
		memberIDs := []uint{aliceID, bobID, carolID}

		expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
		if err != nil {
			t.Fatalf("iter %d create: %v", iter, err)
		}

		// 放大竞态窗口：多个生成与退款同时开始。
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, _ = settleSvc.Generate(aliceID, groupID)
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = expenseSvc.Delete(aliceID, expense.ID)
		}()
		close(start)
		wg.Wait()

		// 不变量：并发结束后，残留的待结算建议金额只可能是 0（退款后生成）
		// 或 ~200（退款前生成、退款随后清空），绝不能在退款提交后仍残留 ~200 的旧建议。
		// 直接重新生成一次以“最新消费数据”为基准：唯一消费已退款，正确结果应为 0。
		finalItems, err := settleSvc.Generate(aliceID, groupID)
		if err != nil {
			t.Fatalf("iter %d final generate: %v", iter, err)
		}
		if len(finalItems) != 0 {
			t.Fatalf("iter %d: after refund, latest-data generate returned %d transfers, want 0",
				iter, len(finalItems))
		}
		total := pendingTransferTotal(t, settleSvc, groupID, aliceID)
		if total > 0.0001 {
			t.Fatalf("iter %d: stale pending suggestions total %.2f after refund, want 0", iter, total)
		}
		for _, uid := range memberIDs {
			pending, err := settleSvc.ListPending(uid)
			if err != nil {
				t.Fatalf("iter %d list pending user %d: %v", iter, uid, err)
			}
			if len(pending) != 0 {
				t.Fatalf("iter %d: user %d has %d stale reminders after refund", iter, uid, len(pending))
			}
		}
	}
}

// TestRefundConcurrentGeneratePartial 多笔消费退款一笔：最终建议必须按剩余有效消费计算。
func TestRefundConcurrentGeneratePartial(t *testing.T) {
	for iter := 0; iter < 10; iter++ {
		db, expenseSvc, settleSvc, groupID, aliceID, bobID, carolID := newConcurrentFixture(t)
		_ = db

		first, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
		if err != nil {
			t.Fatalf("iter %d create first: %v", iter, err)
		}
		second := createExpenseReq(groupID, aliceID, bobID, carolID)
		second.Title = "打车"
		second.Amount = 90
		second.PaidAt = "2026-08-02 12:00:00"
		if _, err := expenseSvc.Create(aliceID, second); err != nil {
			t.Fatalf("iter %d create second: %v", iter, err)
		}

		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, _ = settleSvc.Generate(aliceID, groupID)
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = expenseSvc.Delete(aliceID, first.ID)
		}()
		close(start)
		wg.Wait()

		// 以最新消费数据为基准重新生成：300 那笔已退款，只剩 90，各人应付 30，Alice 应收 60。
		items, err := settleSvc.Generate(aliceID, groupID)
		if err != nil {
			t.Fatalf("iter %d final generate: %v", iter, err)
		}
		total := sumTransferAmounts(items)
		if total < 59.99 || total > 60.01 {
			t.Fatalf("iter %d: latest-data total = %.2f, want ~60 (stale pre-refund data?)", iter, total)
		}
		got := pendingTransferTotal(t, settleSvc, groupID, aliceID)
		if got < 59.99 || got > 60.01 {
			t.Fatalf("iter %d: stored pending total = %.2f, want ~60", iter, got)
		}
	}
}

// TestDuplicateRefundConcurrentGenerate 重复退款与生成并发：只有一次退款成功，最终无过期建议。
func TestDuplicateRefundConcurrentGenerate(t *testing.T) {
	db, expenseSvc, settleSvc, groupID, aliceID, bobID, carolID := newConcurrentFixture(t)
	_ = db
	memberIDs := []uint{aliceID, bobID, carolID}

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	refundErrs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			refundErrs <- expenseSvc.Delete(aliceID, expense.ID)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		_, _ = settleSvc.Generate(aliceID, groupID)
	}()
	close(start)
	wg.Wait()
	close(refundErrs)

	successes, conflicts := 0, 0
	for e := range refundErrs {
		switch {
		case e == nil:
			successes++
		default:
			conflicts++
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("refund outcomes: success=%d conflict=%d, want 1/1", successes, conflicts)
	}

	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("final generate: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("transfers after double refund = %d, want 0", len(items))
	}
	for _, uid := range memberIDs {
		pending, err := settleSvc.ListPending(uid)
		if err != nil {
			t.Fatalf("list pending user %d: %v", uid, err)
		}
		if len(pending) != 0 {
			t.Fatalf("user %d has %d stale reminders", uid, len(pending))
		}
	}
}

// TestGenerateReadLatestAfterRefund 顺序基线：先生成再退款再生成，结果按退款后数据（与并发用例对照）。
func TestGenerateReadLatestAfterRefund(t *testing.T) {
	db, expenseSvc, settleSvc, groupID, aliceID, bobID, carolID := newConcurrentFixture(t)
	_ = db

	expense, err := expenseSvc.Create(aliceID, createExpenseReq(groupID, aliceID, bobID, carolID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := settleSvc.Generate(aliceID, groupID); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if err := expenseSvc.Delete(aliceID, expense.ID); err != nil {
		t.Fatalf("refund: %v", err)
	}
	items, err := settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("transfers after refund = %d, want 0", len(items))
	}

	// 再新增一笔消费并生成，确认生成始终读取最新数据。
	if _, err := expenseSvc.Create(aliceID, &dto.CreateExpenseReq{
		GroupID: groupID, Title: "新聚餐", Amount: 90, Category: "dining",
		PayerID: aliceID, SplitType: "equal", PaidAt: "2026-08-03 12:00:00",
		Shares: []dto.ShareInput{{UserID: aliceID}, {UserID: bobID}, {UserID: carolID}},
	}); err != nil {
		t.Fatalf("recreate: %v", err)
	}
	items, err = settleSvc.Generate(aliceID, groupID)
	if err != nil {
		t.Fatalf("generate after new expense: %v", err)
	}
	if total := sumTransferAmounts(items); total < 59.99 || total > 60.01 {
		t.Fatalf("total after new expense = %.2f, want ~60", total)
	}
}
