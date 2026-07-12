package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	docs "github.com/romeokeita231/ai-router/docs"

	"github.com/romeokeita231/ai-router/internal/config"
	"github.com/romeokeita231/ai-router/internal/controller"
	"github.com/romeokeita231/ai-router/internal/middleware"
	"github.com/romeokeita231/ai-router/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(
	cfg *config.Config,
	healthController *controller.HealthController,
	userController *controller.UserController,
	apiKeyController *controller.ApiKeyController,
	statsController *controller.StatsController,
	internalChatController *controller.InternalChatController,
	chatController *controller.ChatController,
	userService *service.UserService,
	providerController *controller.ProviderController,
	modelController *controller.ModelController,
) (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())

	// 配置 Redis Session
	store, err := redisStore.NewStoreWithDB(
		10, "tcp",
		cfg.RedisAddr, cfg.RedisUsername, cfg.RedisPassword,
		strconv.Itoa(cfg.RedisDB),
		[]byte(cfg.SessionSecret),
	)
	if err != nil {
		return nil, err
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   cfg.SessionMaxAge,
		HttpOnly: true,
	})
	engine.Use(sessions.Sessions(cfg.SessionName, store))

	// 注册路由
	apiGroup := engine.Group(cfg.ContextPath)
	{
		// OpenAPI 文档
		apiGroup.GET("/v3/api-docs", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
		})
		apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(fmt.Sprintf("%s/v3/api-docs", cfg.ContextPath))))

		healthGroup := apiGroup.Group("/health")
		healthGroup.GET("/", healthController.HealthCheck)

		userGroup := apiGroup.Group("/user")
		userGroup.POST("/register", userController.UserRegister)
		userGroup.POST("/login", userController.UserLogin)

		userGroup.POST("/add", middleware.RequireAdmin(userService), userController.AddUser)
		userGroup.POST("/delete", middleware.RequireAdmin(userService), userController.DeleteUser)

		// API Key 管理（需要登录）
		apiKeyGroup := apiGroup.Group("/api/key")
		apiKeyGroup.Use(middleware.RequireLogin(userService))
		apiKeyGroup.POST("/create", apiKeyController.CreateApiKey)
		apiKeyGroup.GET("/list/my", apiKeyController.ListMyApiKeys)
		apiKeyGroup.POST("/revoke", apiKeyController.RevokeApiKey)

		// 统计接口（需要登录）
		statsGroup := apiGroup.Group("/stats")
		statsGroup.Use(middleware.RequireLogin(userService))
		statsGroup.GET("/my/tokens", statsController.GetMyTokenStats)
		statsGroup.GET("/my/logs", statsController.GetMyLogs)

		// 内部对话接口（需要登录）
		internalChatGroup := apiGroup.Group("/internal/chat")
		internalChatGroup.Use(middleware.RequireLogin(userService))
		internalChatGroup.POST("/completions", internalChatController.ChatCompletions)

		// 外部对话接口（通过 API Key 认证，不需要 Session）
		chatGroup := apiGroup.Group("/v1/chat")
		chatGroup.POST("/completions", chatController.ChatCompletions)

		providerGroup := apiGroup.Group("/provider")
		providerGroup.POST("/add", middleware.RequireAdmin(userService), providerController.AddProvider)
		providerGroup.POST("/delete", middleware.RequireAdmin(userService), providerController.DeleteProvider)
		providerGroup.POST("/update", middleware.RequireAdmin(userService), providerController.UpdateProvider)
		providerGroup.GET("/get/vo", providerController.GetProviderVOByID)
		providerGroup.POST("/list/page/vo", providerController.ListProviderVOByPage)
		providerGroup.GET("/list/vo", providerController.ListProviderVO)
		providerGroup.GET("/list/healthy", providerController.ListHealthyProviders)

		modelGroup := apiGroup.Group("/model")
		modelGroup.POST("/add", middleware.RequireAdmin(userService), modelController.AddModel)
		modelGroup.POST("/delete", middleware.RequireAdmin(userService), modelController.DeleteModel)
		modelGroup.POST("/update", middleware.RequireAdmin(userService), modelController.UpdateModel)
		modelGroup.GET("/get/vo", modelController.GetModelVOByID)
		modelGroup.POST("/list/page/vo", modelController.ListModelVOByPage)
		modelGroup.GET("/list/vo", modelController.ListModelVO)
		modelGroup.GET("/list/active", modelController.ListActiveModels)
		modelGroup.GET("/list/active/provider/:providerId", modelController.ListActiveModelsByProvider)
		modelGroup.GET("/list/active/type/:modelType", modelController.ListActiveModelsByType)

	}

	return engine, nil
}
