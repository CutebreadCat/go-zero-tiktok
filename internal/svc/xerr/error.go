package xerr

import "fmt"

// CodeError 自定义错误结构
type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// 实现 error 接口
func (e *CodeError) Error() string {
	return fmt.Sprintf("Code: %d, Msg: %s", e.Code, e.Msg)
}

// 快捷构造函数
func New(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

// 预定义一些常用错误
var (
	ErrInvalidParam  = New(400, "请求参数错误")
	ErrUnauthorized  = New(401, "用户未登录")
	ErrUserNotFound  = New(1001, "用户不存在")
	ErrMysqlError    = New(1002, "数据库错误")
	ErrRedisError    = New(1003, "缓存错误")
	ErrInternalError = New(1004, "服务器内部错误")
	ErrNotFound      = New(404, "资源未找到")
)

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // omitempty 表示如果 data 为空则不显示该字段
}

// 辅助函数：将 CodeError 转换为 ApiResponse
func (e *CodeError) ToResponse() *ApiResponse {
	return &ApiResponse{
		Code: e.Code,
		Msg:  e.Msg,
	}
}
