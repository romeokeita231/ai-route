// @title AI Router API
// @version 1.0
// @description Go backend API 文档
// @BasePath /api
package main



import (
    "log"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"

    "github.com/romeokeita231/ai-router/internal/config"
    "github.com/romeokeita231/ai-router/internal/controller"
    "github.com/romeokeita231/ai-router/internal/repository"
    "github.com/romeokeita231/ai-router/internal/router"
    "github.com/romeokeita231/ai-router/internal/service"
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
    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    healthController := controller.NewHealthController()
    userController := controller.NewUserController(userService)

    // 4. 构建路由并启动服务
    engine, err := router.New(cfg, healthController, userController, userService)
    if err != nil {
        log.Fatalf("build router failed: %v", err)
    }

    if err = engine.Run(":" + cfg.ServerPort); err != nil {
        log.Fatalf("server run failed: %v", err)
    }
}
