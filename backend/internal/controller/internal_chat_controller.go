package controller

import (
	"log"
	"net/http"
	
	
	"github.com/gin-gonic/gin"
	"github.com/romeokeita231/ai-router/internal/common"
	"github.com/romeokeita231/ai-router/internal/service"
	"github.com/romeokeita231/ai-router/internal/errno"
	"github.com/romeokeita231/ai-router/internal/model/dto"
)

type InternalChatController struct {
	chatService *service.ChatService
	userService *service.UserService
}

func NewInternalChatController(chatService *service.ChatService, apiKeyService *service.ApiKeyService, userService *service.UserService) *InternalChatController {
	return &InternalChatController{
		chatService: chatService,
		userService: userService,
	}
}

func (c *InternalChatController) ChatCompletions(ctx *gin.Context) {
	var request dto.ChatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("internal chat bind request failed: path=%s err=%v", ctx.Request.URL.Path, err)
		common.Error(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
		return
	}
	if len(request.Messages) == 0 {
		common.Error(ctx, errno.ParamsError.Code, "messages 不能为空")
		return
	}
	loginUser, err := c.userService.GetLoginUser(ctx)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	apiKeyID := int64(0)
	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	if request.Stream != nil && *request.Stream {
		c.stream(ctx, request, loginUser.ID, apiKeyID, clientIP, userAgent)
		return
	}
	response, err := c.chatService.Chat(request, loginUser.ID, apiKeyID, clientIP, userAgent)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	common.Success(ctx, response)
}



func (c *InternalChatController) stream(ctx *gin.Context, request dto.ChatRequest, userID, apiKeyID int64, clientIP, userAgent string) {
	streamChan, errChan := c.chatService.ChatStream(request, userID, apiKeyID)
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Status(http.StatusOK)

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		log.Printf("internal chat stream not supported by writer")
		common.Error(ctx, errno.SystemError.Code, "流式响应不支持")
		return
	}

	for {
		select {
		case chunk, open := <-streamChan:
			if !open {
				_, _ = ctx.Writer.WriteString("data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			_, _ = ctx.Writer.WriteString("data: " + chunk + "\n\n")
			flusher.Flush()
		case err, open := <-errChan:
			if open && err != nil {
				log.Printf("internal chat stream error: userId=%d apiKeyId=%d err=%v", userID, apiKeyID, err)
				return
			}
		case <-ctx.Request.Context().Done():
			return
		}
	}
}