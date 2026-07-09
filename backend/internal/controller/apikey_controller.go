package controller

import (
    "log"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/romeokeita231/ai-router/internal/model/dto"
	"github.com/romeokeita231/ai-router/internal/model/vo"
	"github.com/romeokeita231/ai-router/internal/service"
	"github.com/romeokeita231/ai-router/internal/errno"
	"github.com/romeokeita231/ai-router/internal/common"
)

type ApiKeyController struct {
    apiKeyService *service.ApiKeyService
    userService   *service.UserService
}

func NewApiKeyController(apiKeyService *service.ApiKeyService, userService *service.UserService) *ApiKeyController {
	return &ApiKeyController{
		apiKeyService: apiKeyService,
		userService:   userService,
	}
}

func (c *ApiKeyController) CreateApiKey(ctx *gin.Context) {
    var request dto.ApiKeyCreateRequest
    if err := ctx.ShouldBindJSON(&request); err != nil {
        common.Error(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
        return
    }
    loginUser, err := c.userService.GetLoginUser(ctx)
    if err != nil {
        c.handleError(ctx, err)
        return
    }
    apiKey, err := c.apiKeyService.CreateApiKey(request.KeyName, loginUser)
    if err != nil {
        c.handleError(ctx, err)
        return
    }
    // 创建时返回完整 Key 值
    common.Success(ctx, vo.ApiKeyVO{
        ID:           apiKey.ID,
        KeyValue:     apiKey.KeyValue,
        KeyName:      apiKey.KeyName,
        Status:       apiKey.Status,
        TotalTokens:  apiKey.TotalTokens,
        LastUsedTime: apiKey.LastUsedTime,
        CreateTime:   apiKey.CreateTime,
    })
}

func (c *ApiKeyController) ListMyApiKeys(ctx *gin.Context) {
	loginUser, err := c.userService.GetLoginUser(ctx)
	if err != nil {
		c.handleError(ctx, err)
		return
	}
	pageNum := parsePositiveInt64WithDefault(ctx.Query("pageNum"), 1)
	pageSize := parsePositiveInt64WithDefault(ctx.Query("pageSize"), 10)
	pageData, err := c.apiKeyService.ListMyApiKeys(loginUser.ID, pageNum, pageSize)
	if err != nil {
		c.handleError(ctx, err)
		return
	}
	common.Success(ctx, pageData)
}

func (c *ApiKeyController) RevokeApiKey(ctx *gin.Context) {
	var request dto.DeleteRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("revoke api key bind failed: %v", err)
		common.Error(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
		return
	}
	if request.ID == nil || request.ID.Int64() <= 0 {
		common.Error(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
		return
	}
	loginUser, err := c.userService.GetLoginUser(ctx)
	if err != nil {
		c.handleError(ctx, err)
		return
	}
	ok, err := c.apiKeyService.RevokeApiKey(request.ID.Int64(), loginUser.ID)
	if err != nil {
		c.handleError(ctx, err)
		return
	}
	common.Success(ctx, ok)
}

func (c *ApiKeyController) handleError(ctx *gin.Context, err error) {
	if bizErr, ok := errno.AsBusinessError(err); ok {
		common.Error(ctx, bizErr.Code, bizErr.Message)
		return
	}
	common.Error(ctx, errno.SystemError.Code, "系统错误")
}

func parsePositiveInt64WithDefault(raw string, defaultValue int64) int64 {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}
