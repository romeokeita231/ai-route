package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/romeokeita231/ai-router/internal/common"
)

type HealthController struct{}

func NewHealthController() *HealthController {
    return &HealthController{}
}

func (h *HealthController) HealthCheck(c *gin.Context) {
    common.Success(c, "ok")
}
