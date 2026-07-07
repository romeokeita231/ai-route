package controller

import (
	"log"
	"strconv"
	"github.com/gin-gonic/gin"
	
	"github.com/romeokeita231/ai-router/internal/model/dto"
	_"github.com/romeokeita231/ai-router/internal/model/vo"
	"github.com/romeokeita231/ai-router/internal/service"
	"github.com/romeokeita231/ai-router/internal/common"
	"github.com/romeokeita231/ai-router/internal/errno"
)

type UserController struct {
    userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
    return &UserController{userService: userService}
}

// @Summary 用户注册
// @Tags userController
// @Accept json
// @Produce json
// @Param request body dto.UserRegisterRequest true "用户注册请求"
// @Success 200 {object} common.BaseResponse{data=string}
// @Router /user/register [post]
// @ID UserRegister
func (u *UserController) UserRegister(c *gin.Context) {
    var request dto.UserRegisterRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        u.handleBindError(c, "user register", err)
        return
    }
    userID, err := u.userService.UserRegister(
        request.UserAccount, request.UserPassword, request.CheckPassword,
    )
    if err != nil {
        u.handleError(c, err)
        return
    }
    common.Success(c, strconv.FormatInt(userID, 10))
}

// @Summary 用户登录
// @Tags userController
// @Accept json
// @Produce json
// @Param request body dto.UserLoginRequest true "用户登录请求"
// @Success 200 {object} common.BaseResponse{data=vo.LoginUserVO}
// @Router /user/login [post]
// @ID UserLogin
func (u *UserController) UserLogin(c *gin.Context) {
    var request dto.UserLoginRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        u.handleBindError(c, "user login", err)
        return
    }
    loginUserVO, err := u.userService.UserLogin(
        request.UserAccount, request.UserPassword, c,
    )
    if err != nil {
        u.handleError(c, err)
        return
    }
    common.Success(c, loginUserVO)
}

// @Summary 添加用户
// @Tags userController
// @Accept json
// @Produce json
// @Param request body dto.UserAddRequest true "添加用户请求"
// @Success 200 {object} common.BaseResponse{data=int64}
// @Router /user/add [post]
// @ID AddUser
func (u *UserController) AddUser(c *gin.Context) {
	var request dto.UserAddRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		u.handleBindError(c, "add user", err)
		return
	}
	userID, err := u.userService.CreateUser(request)
	if err != nil {
		u.handleError(c, err)
		return
	}
	common.Success(c, userID)
}

// @Summary 删除用户
// @Tags userController
// @Accept json
// @Produce json
// @Param request body dto.DeleteRequest true "删除用户请求"
// @Success 200 {object} common.BaseResponse{data=bool}
// @Router /user/delete [post]
// @ID DeleteUser
func (u *UserController) DeleteUser(c *gin.Context) {
	var request dto.DeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		u.handleBindError(c, "delete user", err)
		return
	}
	if request.ID == nil || request.ID.Int64() <= 0 {
		log.Printf("delete user params invalid: id=%v", request.ID)
		common.Error(c, errno.ParamsError.Code, errno.ParamsError.Message)
		return
	}
	result, err := u.userService.DeleteUser(request.ID.Int64())
	if err != nil {
		u.handleError(c, err)
		return
	}
	common.Success(c, result)
}

// @Summary 获取登录用户信息
// @Tags userController
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse{data=vo.LoginUserVO}
// @Router /user/login [get]
// @ID GetLoginUser
func (u *UserController) GetLoginUser(c *gin.Context) {
	loginUser, err := u.userService.GetLoginUser(c)
	if err != nil {
		u.handleError(c, err)
		return
	}
	loginUserVO := u.userService.GetLoginUserVO(loginUser)
	common.Success(c, loginUserVO)
}

// @Summary 用户退出登录
// @Tags userController
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse{data=bool}
// @Router /user/logout [post]
// @ID UserLogout
func (u *UserController) UserLogout(c *gin.Context) {
	if err := u.userService.UserLogout(c); err != nil {
		u.handleError(c, err)
		return
	}
	common.Success(c, true)
}

func (u *UserController) handleError(c *gin.Context, err error) {
    if bizErr, ok := errno.AsBusinessError(err); ok {
        common.Error(c, bizErr.Code, bizErr.Message)
        return
    }
    common.Error(c, errno.SystemError.Code, "系统错误")
}

func (u *UserController) handleBindError(c *gin.Context, action string, err error) {
	log.Printf("bind request failed: action=%s method=%s path=%s err=%v", action, c.Request.Method, c.Request.URL.Path, err)
	common.Error(c, errno.ParamsError.Code, errno.ParamsError.Message)
}

