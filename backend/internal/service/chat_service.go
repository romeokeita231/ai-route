package service

import (
	"io"
	"log"
    "time"
	"bytes"
	"bufio"
	"strings"
    "net/http"
	"encoding/json"
    "github.com/romeokeita231/ai-router/internal/config"
	"github.com/romeokeita231/ai-router/internal/errno"
    "github.com/romeokeita231/ai-router/internal/model/dto"
)

const (
	httpTimeoutSeconds = 60
	chatPath           = "/v1/chat/completions"
)

type upstreamChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type upstreamStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatService struct {
    cfg               *config.Config
    requestLogService *RequestLogService
    httpClient        *http.Client
}

func NewChatService(cfg *config.Config, requestLogService *RequestLogService) *ChatService {
    return &ChatService{
        cfg:               cfg,
        requestLogService: requestLogService,
        httpClient: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

func buildChatPayload(request dto.ChatRequest, modelName string, stream bool) map[string]any {
    payload := map[string]any{
        "model":    modelName,
        "messages": request.Messages,
        "stream":   stream,
    }
    if request.Temperature != nil {
        payload["temperature"] = *request.Temperature
    }
    if request.MaxTokens != nil {
        payload["max_tokens"] = *request.MaxTokens
    }
    if stream {
        payload["stream_options"] = map[string]any{
            "include_usage": true,
        }
    }
    return payload
}

func (s *ChatService) Chat(chatRequest dto.ChatRequest, userID, apiKeyID int64, clientIP, userAgent string) (*dto.ChatResponse, error) {
    start := time.Now()
    modelName := s.ensureModel(chatRequest.Model)

    // 1. 构建请求参数
    payload := buildChatPayload(chatRequest, modelName, false)
    
    // 2. 调用上游模型
    body, err := s.callUpstream(payload)
    if err != nil {
        s.requestLogService.LogRequestAsync(
            ptrInt64(userID), ptrInt64(apiKeyID), modelName,
            0, 0, 0, int(time.Since(start).Milliseconds()), "failed", err.Error(),
        )
        return nil, errno.NewWithMessage(errno.SystemError, "调用模型失败: "+err.Error())
    }

    // 3. 解析上游响应
    var upstreamResp upstreamChatResponse
    if err = json.Unmarshal(body, &upstreamResp); err != nil {
        s.requestLogService.LogRequestAsync(
            ptrInt64(userID), ptrInt64(apiKeyID), modelName,
            0, 0, 0, int(time.Since(start).Milliseconds()), "failed", err.Error(),
        )
        return nil, errno.NewWithMessage(errno.SystemError, "调用模型失败: "+err.Error())
    }

    // 4. 转换为标准响应格式
    response := mapChatResponse(upstreamResp, modelName)
    
    // 5. 异步记录请求日志
    usage := response.Usage
    s.requestLogService.LogRequestAsync(
        ptrInt64(userID), ptrInt64(apiKeyID), modelName,
        usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens,
        int(time.Since(start).Milliseconds()), "success", "",
    )
    return &response, nil
}

func (s *ChatService) callUpstream(payload map[string]any) ([]byte, error) {
    // 检查 API Key 是否配置
    if strings.TrimSpace(s.cfg.AIAPIKey) == "" || 
       strings.Contains(s.cfg.AIAPIKey, "YOUR_QWEN_API_KEY") {
        return nil, errno.NewWithMessage(errno.SystemError, "AI_API_KEY 未配置")
    }
    rawPayload, _ := json.Marshal(payload)
    
    // 构造 HTTP 请求
    req, _ := http.NewRequest(
        http.MethodPost,
        strings.TrimRight(s.cfg.AIBaseURL, "/")+"/v1/chat/completions",
        bytes.NewReader(rawPayload),
    )
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+s.cfg.AIAPIKey)

    // 发送请求
    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    responseBody, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= http.StatusBadRequest {
        return nil, errno.NewWithMessage(errno.SystemError, string(responseBody))
    }
    return responseBody, nil
}

func mapChatResponse(upstream upstreamChatResponse, defaultModel string) dto.ChatResponse {
    modelName := upstream.Model
    if modelName == "" {
        modelName = defaultModel
    }
    choices := make([]dto.ChatResponseChoice, 0, len(upstream.Choices))
    for _, item := range upstream.Choices {
        choices = append(choices, dto.ChatResponseChoice{
            Index: item.Index,
            Message: dto.ChatMessage{
                Role:    item.Message.Role,
                Content: item.Message.Content,
            },
            FinishReason: item.FinishReason,
        })
    }
    return dto.ChatResponse{
        ID:      upstream.ID,
        Object:  "chat.completion",
        Created: upstream.Created,
        Model:   modelName,
        Choices: choices,
        Usage: dto.ChatResponseUsage{
            PromptTokens:     upstream.Usage.PromptTokens,
            CompletionTokens: upstream.Usage.CompletionTokens,
            TotalTokens:      upstream.Usage.TotalTokens,
        },
    }
}

func (s *ChatService) ChatStream(chatRequest dto.ChatRequest, userID, apiKeyID int64) (<-chan string, <-chan error) {
    streamChan := make(chan string, 32)
    errChan := make(chan error, 1)

    go func() {
        defer close(streamChan)
        defer close(errChan)

        start := time.Now()
        modelName := s.ensureModel(chatRequest.Model)
        payload := buildChatPayload(chatRequest, modelName, true)
        
        // 1. 建立流式连接
        resp, err := s.callUpstreamStream(payload)
        if err != nil {
            s.requestLogService.LogRequestAsync(
                ptrInt64(userID), ptrInt64(apiKeyID), modelName,
                0, 0, 0, int(time.Since(start).Milliseconds()), "failed", err.Error(),
            )
            errChan <- err
            return
        }
        defer resp.Body.Close()

        // 2. 逐行读取 SSE 数据
        promptTokens := 0
        completionTokens := 0
        totalTokens := 0
        reader := bufio.NewReader(resp.Body)

        for {
            line, readErr := reader.ReadString('\n')
            if readErr != nil {
                if readErr == io.EOF {
                    break
                }
                errChan <- readErr
                return
            }
            line = strings.TrimSpace(line)
            if line == "" || !strings.HasPrefix(line, "data:") {
                continue
            }
            rawData := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
            if rawData == "[DONE]" {
                break
            }

            // 3. 解析每个数据块
            var chunk upstreamStreamChunk
            if err = json.Unmarshal([]byte(rawData), &chunk); err != nil {
                continue
            }
            
            // 4. 收集 Token 统计（通常只有最后一个 chunk 才有）
            if chunk.Usage.PromptTokens > 0 {
                promptTokens = chunk.Usage.PromptTokens
            }
            if chunk.Usage.CompletionTokens > 0 {
                completionTokens = chunk.Usage.CompletionTokens
            }
            if chunk.Usage.TotalTokens > 0 {
                totalTokens = chunk.Usage.TotalTokens
            }
            
            if len(chunk.Choices) == 0 {
                continue
            }
            content := chunk.Choices[0].Delta.Content
            if content == "" {
                continue
            }
            
            // 5. 转义换行符并推送到 Channel
            streamChan <- strings.ReplaceAll(content, "\n", "\\n")
        }

        // 6. 流结束，记录日志
        if totalTokens == 0 {
            totalTokens = promptTokens + completionTokens
        }
        s.requestLogService.LogRequestAsync(
            ptrInt64(userID), ptrInt64(apiKeyID), modelName,
            promptTokens, completionTokens, totalTokens,
            int(time.Since(start).Milliseconds()), "success", "",
        )
    }()

    return streamChan, errChan
}

func (s *ChatService) callUpstreamStream(payload map[string]any) (*http.Response, error) {
	if strings.TrimSpace(s.cfg.AIAPIKey) == "" || strings.Contains(s.cfg.AIAPIKey, "YOUR_QWEN_API_KEY") {
		return nil, errno.NewWithMessage(errno.SystemError, "AI_API_KEY 未配置")
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(s.cfg.AIBaseURL, "/")+chatPath, bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.AIAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		responseBody, _ := io.ReadAll(resp.Body)
		log.Printf("chat stream upstream bad status: status=%d body=%s", resp.StatusCode, trimForLog(string(responseBody)))
		return nil, errno.NewWithMessage(errno.SystemError, string(responseBody))
	}
	return resp, nil
}

func ptrInt64(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	value := v
	return &value
}

func (s *ChatService) ensureModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return s.cfg.AIModel
	}
	return model
}

func trimForLog(value string) string {
	const maxLogLength = 1200
	if len(value) <= maxLogLength {
		return value
	}
	return value[:maxLogLength] + "...(truncated)"
}

