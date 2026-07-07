package errno

type BusinessError struct {
    Code    int
    Message string
}

func (e *BusinessError) Error() string {
    return e.Message
}

func New(code ErrorCode) *BusinessError {
    return &BusinessError{
        Code:    code.Code,
        Message: code.Message,
    }
}

func NewWithMessage(code ErrorCode, message string) *BusinessError {
    return &BusinessError{
        Code:    code.Code,
        Message: message,
    }
}

// AsBusinessError 尝试将 error 转换为 BusinessError
func AsBusinessError(err error) (*BusinessError, bool) {
    if err == nil {
        return nil, false
    }
    bizErr, ok := err.(*BusinessError)
    return bizErr, ok
}
