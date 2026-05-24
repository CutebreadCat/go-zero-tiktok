package xerr

import (
	"errors"
	"fmt"

	pkgerrors "github.com/pkg/errors"
)

var ErrRedisError = errors.New("redis error")

const (
	ServerBusy   = 1001
	InvalidParam = 1002
	Unauthorized = 401
)

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	err  error
}

func (e *CodeError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("Code: %d, Msg: %s, Err: %v", e.Code, e.Msg, e.err)
	}
	return fmt.Sprintf("Code: %d, Msg: %s", e.Code, e.Msg)
}

func (e *CodeError) Unwrap() error {
	return e.err
}

func (e *CodeError) Cause() error {
	return e.err
}

func New(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

func NewServerBusy() error {
	return &CodeError{Code: ServerBusy, Msg: "系统繁忙"}
}

func WrapServerBusy(err error) error {
	return &CodeError{Code: ServerBusy, Msg: "系统繁忙", err: err}
}

func NewInvalidParam(msg string) error {
	return &CodeError{Code: InvalidParam, Msg: msg}
}

func NewUnauthorized(msg string) error {
	return &CodeError{Code: Unauthorized, Msg: msg}
}

func WithMessage(err error, message string) error {
	return pkgerrors.WithMessage(err, message)
}

func WithMessagef(err error, format string, args ...interface{}) error {
	return pkgerrors.WithMessagef(err, format, args...)
}

func Wrap(err error, message string) error {
	return pkgerrors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return pkgerrors.Wrapf(err, format, args...)
}

func Cause(err error) error {
	return pkgerrors.Cause(err)
}

func HandleDaoError(err error, context string) error {
	var codeErr *CodeError
	if errors.As(err, &codeErr) {
		return err
	}
	return &CodeError{Code: ServerBusy, Msg: "系统繁忙", err: pkgerrors.WithMessage(err, context)}
}

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (e *CodeError) ToResponse() *ApiResponse {
	return &ApiResponse{
		Code: e.Code,
		Msg:  e.Msg,
	}
}

func (e *CodeError) HandleResponse() *ApiResponse {
	switch e.Code {
	case ServerBusy:
		if e.err != nil {
			fmt.Printf("[ServerBusy] code=%d, msg=%s, err=%+v\n", e.Code, e.Msg, e.err)
		}
		return &ApiResponse{
			Code: e.Code,
			Msg:  "系统繁忙",
		}
	case InvalidParam:
		return &ApiResponse{
			Code: e.Code,
			Msg:  e.Msg,
		}
	case Unauthorized:
		return &ApiResponse{
			Code: e.Code,
			Msg:  e.Msg,
		}
	default:
		if e.err != nil {
			fmt.Printf("[InternalError] code=%d, msg=%s, err=%+v\n", e.Code, e.Msg, e.err)
		}
		return &ApiResponse{
			Code: e.Code,
			Msg:  e.Msg,
		}
	}
}

func HandleError(err error) *ApiResponse {
	var codeErr *CodeError
	if errors.As(err, &codeErr) {
		return codeErr.HandleResponse()
	}

	fmt.Printf("[UnknownError] err=%+v\n", err)
	return &ApiResponse{
		Code: ServerBusy,
		Msg:  "系统繁忙",
	}
}
