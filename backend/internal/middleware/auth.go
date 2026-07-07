package middleware

import (
	
	"github.com/gin-gonic/gin"
	
	
	"github.com/romeokeita231/ai-router/internal/service"
	"github.com/romeokeita231/ai-router/internal/common"
	"github.com/romeokeita231/ai-router/internal/constant"
	"github.com/romeokeita231/ai-router/internal/errno"
)

func RequireLogin(userService *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        if _, err := userService.GetLoginUser(c); err != nil {
            if bizErr, ok := errno.AsBusinessError(err); ok {
                common.Error(c, bizErr.Code, bizErr.Message)
            } else {
                common.Error(c, errno.SystemError.Code, "系统错误")
            }
            c.Abort()
            return
        }
        c.Next()
    }
}

func RequireAdmin(userService *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        loginUser, err := userService.GetLoginUser(c)
        if err != nil {
            // ... 错误处理
            c.Abort()
            return
        }
        if loginUser.UserRole != constant.AdminRole {
            common.Error(c, errno.NoAuthError.Code, errno.NoAuthError.Message)
            c.Abort()
            return
        }
        c.Next()
    }
}
