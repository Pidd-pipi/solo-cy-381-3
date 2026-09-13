package router

import (
	"log/slog"

	"github.com/aasplit/aasplit/internal/config"
	"github.com/aasplit/aasplit/internal/handler"
	"github.com/aasplit/aasplit/internal/middleware"
	"github.com/aasplit/aasplit/internal/repository"
	"github.com/aasplit/aasplit/internal/service"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Setup 装配依赖并注册全部路由。
func Setup(cfg *config.Config, db *gorm.DB, rdb *redis.Client, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.ErrorHandler(logger))
	engine.Use(middleware.RequestLogger(logger))
	engine.Use(middleware.CORS(cfg.CORSOrigins))
	engine.Use(middleware.NewRateLimiter(rdb, cfg.RateLimit).Middleware())

	jwtMgr := util.NewJWTManager(cfg.JWTSecret, cfg.JWTExpire())

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	memberRepo := repository.NewGroupMemberRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	shareRepo := repository.NewExpenseShareRepository(db)
	settleRepo := repository.NewSettlementRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	auditSvc := service.NewAuditService(auditRepo, logger)
	userSvc := service.NewUserService(userRepo, jwtMgr, logger)
	groupSvc := service.NewGroupService(db, groupRepo, memberRepo, userRepo, auditSvc, logger)
	expenseSvc := service.NewExpenseService(db, expenseRepo, memberRepo, groupRepo, userRepo, auditSvc, logger)
	settleSvc := service.NewSettlementService(db, settleRepo, shareRepo, memberRepo, groupRepo, userRepo, auditSvc, logger)
	statsSvc := service.NewStatsService(statsRepo, shareRepo, memberRepo, logger)

	userHandler := handler.NewUserHandler(userSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	expenseHandler := handler.NewExpenseHandler(expenseSvc)
	settleHandler := handler.NewSettlementHandler(settleSvc)
	auditHandler := handler.NewAuditHandler(auditSvc, userRepo)
	statsHandler := handler.NewStatsHandler(statsSvc)

	engine.GET("/healthz", func(c *gin.Context) {
		util.OK(c, gin.H{"status": "healthy", "app": cfg.AppName, "env": cfg.Env})
	})

	// 通用 HTTP 审计中间件挂在 /api/v1 写操作上
	v1 := engine.Group("/api/v1", middleware.Audit(auditRepo, logger))
	{
		RegisterUserRoutes(v1, userHandler, jwtMgr)
		RegisterGroupRoutes(v1, groupHandler, jwtMgr)
		RegisterExpenseRoutes(v1, expenseHandler, jwtMgr)
		RegisterSettlementRoutes(v1, settleHandler, jwtMgr)
		RegisterAuditRoutes(v1, auditHandler, jwtMgr)
		RegisterStatsRoutes(v1, statsHandler, jwtMgr)
	}
	engine.NoRoute(middleware.NoRoute)
	engine.NoMethod(middleware.NoMethod)
	return engine
}
