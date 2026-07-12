package strategy

import "github.com/romeokeita231/ai-router/internal/model/entity"

type RoutingStrategy interface {
    SelectModel(models []entity.Model, requestedModel string) *entity.Model
    GetFallbackModels(models []entity.Model, requestedModel string) []entity.Model
    GetStrategyType() string
}
