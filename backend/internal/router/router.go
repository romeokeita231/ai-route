package router

import (
	"strconv"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	docs "github.com/romeokeita231/ai-router/docs"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

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
	userService *service.UserService,
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

	}

	return engine, nil
}
