package controller

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/romeokeita231/ai-router/internal/common"
	"github.com/romeokeita231/ai-router/internal/errno"
	"github.com/romeokeita231/ai-router/internal/service"
)

type BlacklistController struct {
	blacklistService *service.BlacklistService
}

type blacklistRequest struct {
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

func NewBlacklistController(blacklistService *service.BlacklistService) *BlacklistController {
	return &BlacklistController{blacklistService: blacklistService}
}

// @Summary 获取黑名单列表
// @Tags blacklistController
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse{data=[]string}
// @Router /blacklist/list [get]
// @ID getBlacklist
func (c *BlacklistController) List(ctx *gin.Context) {
	data, err := c.blacklistService.ListAll()
	if err != nil {
		log.Printf("list blacklist failed: err=%v", err)
		common.Error(ctx, errno.SystemError.Code, "获取黑名单失败")
		return
	}
	common.Success(ctx, data)
}

// @Summary 添加黑名单
// @Tags blacklistController
// @Accept json
// @Produce json
// @Param req body blacklistRequest true "IP and Reason"
// @Success 200 {object} common.BaseResponse{data=bool}
// @Router /blacklist/add [post]
// @ID addToBlacklist
func (c *BlacklistController) Add(ctx *gin.Context) {
	var req blacklistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Error(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
		return
	}
	if strings.TrimSpace(req.IP) == "" {
		common.Error(ctx, errno.ParamsError.Code, "ip 不能为空")
		return
	}
	if err := c.blacklistService.AddToBlacklist(req.IP, req.Reason); err != nil {
		log.Printf("add blacklist failed: ip=%s err=%v", req.IP, err)
		common.Error(ctx, errno.SystemError.Code, "添加黑名单失败")
		return
	}
	common.Success(ctx, true)
}

// @Summary 移除黑名单
// @Tags blacklistController
// @Accept json
// @Produce json
// @Param req body blacklistRequest true "IP"
// @Success 200 {object} common.BaseResponse{data=bool}
// @Router /blacklist/remove [post]
// @ID removeFromBlacklist
func (c *BlacklistController) Remove(ctx *gin.Context) {
	var req blacklistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Error(ctx, errno.ParamsError.Code, errno.ParamsError.Message)
		return
	}
	if strings.TrimSpace(req.IP) == "" {
		common.Error(ctx, errno.ParamsError.Code, "ip 不能为空")
		return
	}
	if err := c.blacklistService.RemoveFromBlacklist(req.IP); err != nil {
		log.Printf("remove blacklist failed: ip=%s err=%v", req.IP, err)
		common.Error(ctx, errno.SystemError.Code, "移除黑名单失败")
		return
	}
	common.Success(ctx, true)
}

// @Summary 检查是否在黑名单
// @Tags blacklistController
// @Accept json
// @Produce json
// @Param ip query string true "IP"
// @Success 200 {object} common.BaseResponse{data=bool}
// @Router /blacklist/check [get]
// @ID checkBlacklist
func (c *BlacklistController) Check(ctx *gin.Context) {
	ip := strings.TrimSpace(ctx.Query("ip"))
	if ip == "" {
		common.Error(ctx, errno.ParamsError.Code, "ip 不能为空")
		return
	}
	common.Success(ctx, c.blacklistService.IsBlocked(ip))
}

// @Summary 获取黑名单数量
// @Tags blacklistController
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse{data=int64}
// @Router /blacklist/count [get]
// @ID countBlacklist
func (c *BlacklistController) Count(ctx *gin.Context) {
	count, err := c.blacklistService.Count()
	if err != nil {
		log.Printf("count blacklist failed: err=%v", err)
		common.Error(ctx, errno.SystemError.Code, "获取黑名单数量失败")
		return
	}
	common.Success(ctx, count)
}
