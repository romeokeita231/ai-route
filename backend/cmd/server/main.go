// @title AI Router API
// @version 1.0
// @description Go backend API 文档
// @BasePath /api
package main

import (
	"log"
	"context"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/romeokeita231/ai-router/internal/adapter"
	"github.com/romeokeita231/ai-router/internal/config"
	"github.com/romeokeita231/ai-router/internal/controller"
	"github.com/romeokeita231/ai-router/internal/repository"
	"github.com/romeokeita231/ai-router/internal/router"
	"github.com/romeokeita231/ai-router/internal/service"
	"github.com/romeokeita231/ai-router/internal/strategy"
	"github.com/romeokeita231/ai-router/internal/task"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 2. 初始化数据库连接
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("open mysql with gorm failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get sql db failed: %v", err)
	}
	defer sqlDB.Close()

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("ping mysql failed: %v", err)
	}

	// 3. 依赖注入（手动组装）
	// Repository 层
	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	requestLogRepo := repository.NewRequestLogRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	modelRepo := repository.NewModelRepository(db)
	
	// Service 层
	userService := service.NewUserService(userRepo)
	apiKeyService := service.NewApiKeyService(apiKeyRepo)
	requestLogService := service.NewRequestLogService(requestLogRepo, apiKeyService)
	providerService := service.NewProviderService(providerRepo)
	modelService := service.NewModelService(modelRepo, providerRepo)
	healthCheckService := service.NewHealthCheckService(providerRepo, modelRepo, requestLogRepo)

	routingStrategies := []strategy.RoutingStrategy{
		strategy.NewAutoRoutingStrategy(),
		strategy.NewFixedRoutingStrategy(),
		strategy.NewCostFirstRoutingStrategy(),
		strategy.NewLatencyFirstRoutingStrategy(),
	}
	routingService := service.NewRoutingService(modelRepo, routingStrategies)

	adapterFactory := adapter.NewModelAdapterFactory(
		[]adapter.ModelAdapter{
			adapter.NewZhipuAdapter(),
			adapter.NewOpenAIAdapter(),
		},
		adapter.NewDefaultAdapter(),
	)
	modelInvokeService := service.NewModelInvokeService(adapterFactory)
	chatService := service.NewChatService(requestLogService, routingService, modelInvokeService, providerService)

	// Controller 层
	healthController := controller.NewHealthController()
	userController := controller.NewUserController(userService)
	apiKeyController := controller.NewApiKeyController(apiKeyService, userService)
	providerController := controller.NewProviderController(providerService)
	modelController := controller.NewModelController(modelService)
	chatController := controller.NewChatController(chatService, apiKeyService)
	internalChatController := controller.NewInternalChatController(chatService, apiKeyService, userService)
	statsController := controller.NewStatsController(requestLogService, userService)
	healthCheckTask := task.NewHealthCheckTask(healthCheckService)

	// 4. 构建路由并启动服务
	engine, err := router.New(
		cfg,
		healthController,
		userController,
		apiKeyController,
		statsController,
		internalChatController,
		chatController,
		userService,
		providerController,
		modelController,
	)
	if err != nil {
		log.Fatalf("build router failed: %v", err)
	}

	taskCtx, cancelTask := context.WithCancel(context.Background())
	defer cancelTask()
	healthCheckTask.Start(taskCtx)

	if err = engine.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
