package controller

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/romeokeita231/ai-router/internal/model/dto"
	"github.com/romeokeita231/ai-router/internal/service"
	"github.com/romeokeita231/ai-router/internal/errno"
	
)

type ChatController struct {
	chatService   *service.ChatService
	apiKeyService *service.ApiKeyService
}

func NewChatController(chatService *service.ChatService, apiKeyService *service.ApiKeyService) *ChatController {
	return &ChatController{
		chatService:   chatService,
		apiKeyService: apiKeyService,
	}
}
func (c *ChatController) ChatCompletions(ctx *gin.Context) {
    var request dto.ChatRequest
    if err := ctx.ShouldBindJSON(&request); err != nil {
        writeError(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
        return
    }
    if len(request.Messages) == 0 {
        writeError(ctx, errno.ParamsError.Code, "messages 不能为空")
        return
    }
    // 验证 API Key
    authorization := ctx.GetHeader("Authorization")
    if !strings.HasPrefix(authorization, "Bearer ") {
        writeError(ctx, errno.NoAuthError.Code, "缺少或无效的 Authorization Header")
        return
    }
    apiKeyValue := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
    apiKey, err := c.apiKeyService.GetByKeyValue(apiKeyValue)
    if err != nil || apiKey == nil {
        writeError(ctx, errno.NoAuthError.Code, "API Key 无效或已失效")
        return
    }

    // 判断是否为流式请求
    if request.Stream != nil && *request.Stream {
        c.stream(ctx, request, apiKey.UserID, apiKey.ID)
        return
    }
    response, err := c.chatService.Chat(request, apiKey.UserID, apiKey.ID, apiKeyValue, apiKeyValue)
    if err != nil {
        writeServiceError(ctx, err)
        return
    }
    ctx.JSON(http.StatusOK, response)
}

func (c *ChatController) stream(ctx *gin.Context, request dto.ChatRequest, userID, apiKeyID int64) {
    streamChan, errChan := c.chatService.ChatStream(request, userID, apiKeyID)
    
    // 设置 SSE 响应头
    ctx.Header("Content-Type", "text/event-stream")
    ctx.Header("Cache-Control", "no-cache")
    ctx.Header("Connection", "keep-alive")
    ctx.Status(http.StatusOK)

    flusher, ok := ctx.Writer.(http.Flusher)
    if !ok {
        writeError(ctx, errno.SystemError.Code, "流式响应不支持")
        return
    }

    for {
        select {
        case chunk, open := <-streamChan:
            if !open {
                // Channel 关闭，流结束
                _, _ = ctx.Writer.WriteString("data: [DONE]\n\n")
                flusher.Flush()
                return
            }
            // 写入 SSE 格式数据
            _, _ = ctx.Writer.WriteString("data: " + chunk + "\n\n")
            flusher.Flush()
        case err, open := <-errChan:
            if open && err != nil {
                return
            }
        case <-ctx.Request.Context().Done():
            // 客户端断开连接
            return
        }
    }
}
