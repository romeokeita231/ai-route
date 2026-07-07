package service

import (
	"strings"
	"strconv"
	"crypto/md5"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"



	"github.com/romeokeita231/ai-router/internal/model/dto"
	"github.com/romeokeita231/ai-router/internal/model/entity"
	"github.com/romeokeita231/ai-router/internal/model/vo"
	"github.com/romeokeita231/ai-router/internal/repository"
	"github.com/romeokeita231/ai-router/internal/constant"
	"github.com/romeokeita231/ai-router/internal/errno"
)

const (
	minAccountLength  = 4
	minPasswordLength = 8
	unlimitedQuota    = int64(-1)
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) UserRegister(userAccount, userPassword, checkPassword string) (int64, error) {
    // 1. 校验参数
    if hasBlank(userAccount, userPassword, checkPassword) {
        return 0, errno.NewWithMessage(errno.ParamsError, "参数为空")
    }
    if len(userAccount) < minAccountLength {
        return 0, errno.NewWithMessage(errno.ParamsError, "账号长度过短")
    }
    if len(userPassword) < minPasswordLength || len(checkPassword) < minPasswordLength {
        return 0, errno.NewWithMessage(errno.ParamsError, "密码长度过短")
    }
    if userPassword != checkPassword {
        return 0, errno.NewWithMessage(errno.ParamsError, "两次输入的密码不一致")
    }

    // 2. 查询用户是否已存在
    count, err := s.userRepo.CountByAccount(userAccount)
    if err != nil {
        return 0, errno.New(errno.SystemError)
    }
    if count > 0 {
        return 0, errno.NewWithMessage(errno.ParamsError, "账号重复")
    }

    // 3. 加密密码并创建用户
    user := &entity.User{
        UserAccount:  userAccount,
        UserPassword: s.GetEncryptPassword(userPassword),
        UserName:     constant.DefaultUserName,
        UserRole:     constant.DefaultRole,
    }
    id, err := s.userRepo.Create(user)
    if err != nil {
        return 0, errno.NewWithMessage(errno.OperationError, "注册失败，数据库错误")
    }
    return id, nil
}

func (s *UserService) UserLogin(userAccount, userPassword string, c *gin.Context) (*vo.LoginUserVO, error) {
    // 1. 校验参数
    if hasBlank(userAccount, userPassword) {
        return nil, errno.NewWithMessage(errno.ParamsError, "参数为空")
    }

    // 2. 查询用户
    encryptPassword := s.GetEncryptPassword(userPassword)
    user, err := s.userRepo.GetByAccountAndPassword(userAccount, encryptPassword)
    if err != nil {
        return nil, errno.New(errno.SystemError)
    }
    if user == nil {
        return nil, errno.NewWithMessage(errno.ParamsError, "用户不存在或密码错误")
    }

    // 3. 记录登录态到 Session
    session := sessions.Default(c)
    session.Set(constant.UserLoginState, strconv.FormatInt(user.ID, 10))
    if err = session.Save(); err != nil {
        return nil, errno.New(errno.SystemError)
    }

    // 4. 返回脱敏的用户信息
    loginUserVO := s.GetLoginUserVO(user)
    return &loginUserVO, nil
}

func (s *UserService) CreateUser(req dto.UserAddRequest) (int64, error) {
	user := &entity.User{
		UserName:     req.UserName,
		UserAccount:  req.UserAccount,
		UserAvatar:   req.UserAvatar,
		UserProfile:  req.UserProfile,
		UserRole:     req.UserRole,
		UserPassword: s.GetEncryptPassword(constant.DefaultUserPassword),
	}
	id, err := s.userRepo.Create(user)
	if err != nil {
		return 0, errno.New(errno.OperationError)
	}
	return id, nil
}

func (s *UserService) DeleteUser(id int64) (bool, error) {
	ok, err := s.userRepo.SoftDeleteByID(id)
	if err != nil {
		return false, errno.New(errno.SystemError)
	}
	return ok, nil
}

func (s *UserService) GetLoginUser(c *gin.Context) (*entity.User, error) {
	session := sessions.Default(c)
	sessionValue := session.Get(constant.UserLoginState)
	if sessionValue == nil {
		return nil, errno.New(errno.NotLoginError)
	}

	userIDStr, ok := sessionValue.(string)
	if !ok || userIDStr == "" {
		return nil, errno.New(errno.NotLoginError)
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		return nil, errno.New(errno.NotLoginError)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errno.New(errno.SystemError)
	}
	if user == nil {
		return nil, errno.New(errno.NotLoginError)
	}
	return user, nil
}

func (s *UserService) GetLoginUserVO(user *entity.User) vo.LoginUserVO {
	return vo.LoginUserVO{
		ID:          user.ID,
		UserAccount: user.UserAccount,
		UserName:    user.UserName,
		UserAvatar:  user.UserAvatar,
		UserProfile: user.UserProfile,
		UserRole:    user.UserRole,
		CreateTime:  user.CreateTime,
		UpdateTime:  user.UpdateTime,
	}
}

func (s *UserService) UserLogout(c *gin.Context) error {
	session := sessions.Default(c)
	sessionValue := session.Get(constant.UserLoginState)
	if sessionValue == nil {
		return errno.NewWithMessage(errno.OperationError, "用户未登录")
	}
	session.Delete(constant.UserLoginState)
	if err := session.Save(); err != nil {
		return errno.New(errno.SystemError)
	}
	return nil
}

func (s *UserService) GetEncryptPassword(userPassword string) string {
    sum := md5.Sum([]byte(userPassword + constant.PasswordSalt))
    return hex.EncodeToString(sum[:])
}

func hasBlank(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}