package dto

import (
    "encoding/json"
    "strconv"
    "strings"
)

type UserRegisterRequest struct {
    UserAccount   string `json:"userAccount"`
    UserPassword  string `json:"userPassword"`
    CheckPassword string `json:"checkPassword"`
}

type UserLoginRequest struct {
    UserAccount  string `json:"userAccount"`
    UserPassword string `json:"userPassword"`
}

type UserAddRequest struct {
    UserName    string `json:"userName"`
    UserAccount string `json:"userAccount"`
    UserAvatar  string `json:"userAvatar"`
    UserProfile string `json:"userProfile"`
    UserRole    string `json:"userRole"`
}

type UserUpdateRequest struct {
    ID          *FlexibleInt64 `json:"id"`
    UserName    string         `json:"userName"`
    UserAvatar  string         `json:"userAvatar"`
    UserProfile string         `json:"userProfile"`
    UserRole    string         `json:"userRole"`
}

type DeleteRequest struct {
    ID *FlexibleInt64 `json:"id"`
}

type UserQueryRequest struct {
    PageNum     int64          `json:"pageNum"`
    PageSize    int64          `json:"pageSize"`
    SortField   string         `json:"sortField"`
    SortOrder   string         `json:"sortOrder"`
    ID          *FlexibleInt64 `json:"id"`
    UserName    string         `json:"userName"`
    UserAccount string         `json:"userAccount"`
    UserProfile string         `json:"userProfile"`
    UserRole    string         `json:"userRole"`
}

type FlexibleInt64 int64

func (v *FlexibleInt64) UnmarshalJSON(data []byte) error {
    trimmed := strings.TrimSpace(string(data))
    if trimmed == "" || trimmed == "null" {
        return nil
    }
    // 如果是字符串格式，先去掉引号再解析
    if strings.HasPrefix(trimmed, "\"") {
        var text string
        if err := json.Unmarshal(data, &text); err != nil {
            return err
        }
        parsed, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
        if err != nil {
            return err
        }
        *v = FlexibleInt64(parsed)
        return nil
    }
    // 直接解析数字
    parsed, err := strconv.ParseInt(trimmed, 10, 64)
    if err != nil {
        return err
    }
    *v = FlexibleInt64(parsed)
    return nil
}

func (v FlexibleInt64) Int64() int64 {
	return int64(v)
}