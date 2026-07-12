package adapter

import (
	"net/http"

	"github.com/romeokeita231/ai-router/internal/model/dto"
	"github.com/romeokeita231/ai-router/internal/model/entity"
)

type ModelAdapter interface {
	Supports(providerName string) bool
	Invoke(model *entity.Model, provider *entity.ModelProvider, chatRequest dto.ChatRequest) ([]byte, error)
	InvokeStream(model *entity.Model, provider *entity.ModelProvider, chatRequest dto.ChatRequest) (*http.Response, error)
}
