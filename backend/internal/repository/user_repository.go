package repository

import (
	"strings"
    "gorm.io/gorm"
	"github.com/romeokeita231/ai-router/internal/model/entity"
	"github.com/romeokeita231/ai-router/internal/model/dto"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) DB() *gorm.DB {
	return r.db
}

func (r *UserRepository) baseQuery() *gorm.DB {
	return r.db.Model(&entity.User{})
}

func (r *UserRepository) Create(user *entity.User) (int64, error) {
    if err := r.db.
        Select("userAccount", "userPassword", "userName", "userAvatar", "userProfile", "userRole").
        Create(user).Error; err != nil {
        return 0, err
    }
    return user.ID, nil
}

func (r *UserRepository) CountByAccount(userAccount string) (int64, error) {
    var count int64
    err := r.baseQuery().
        Where("userAccount = ?", userAccount).
        Count(&count).Error
    return count, err
}

func (r *UserRepository) GetByAccountAndPassword(userAccount, userPassword string) (*entity.User, error) {
    var user entity.User
    err := r.baseQuery().
        Where("userAccount = ? AND userPassword = ?", userAccount, userPassword).
        Take(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) SoftDeleteByID(id int64) (bool, error) {
	result := r.baseQuery().
		Where("id = ?", id).
		Delete(&entity.User{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *UserRepository) ListByPage(req dto.UserQueryRequest) ([]entity.User, int64, error) {
    query := r.baseQuery()

    // 动态构建查询条件
    if req.ID != nil && *req.ID > 0 {
        query = query.Where("id = ?", req.ID.Int64())
    }
    if strings.TrimSpace(req.UserName) != "" {
        query = query.Where("userName LIKE ?", "%"+req.UserName+"%")
    }
    // ... 其他条件

    // 统计总数
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 排序和分页
    offset := (req.PageNum - 1) * req.PageSize
    users := make([]entity.User, 0)
    err := query.Offset(int(offset)).Limit(int(req.PageSize)).Find(&users).Error
    return users, total, err
}

func (r *UserRepository) GetByID(id int64) (*entity.User, error) {
	var user entity.User
	err := r.baseQuery().Where("id = ?", id).Take(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}