package xerr

import (
	"errors"
	"strings"
	"testing"
)

func TestCodeError(t *testing.T) {
	err := New(123, "custom")
	codeErr, ok := err.(*CodeError)
	if !ok {
		t.Fatalf("New returned %T, want *CodeError", err)
	}
	if codeErr.Code != 123 || codeErr.Msg != "custom" {
		t.Fatalf("code error = %+v", codeErr)
	}
	if got := codeErr.Error(); !strings.Contains(got, "Code: 123") || !strings.Contains(got, "custom") {
		t.Fatalf("Error() = %q", got)
	}
	if codeErr.Unwrap() != nil || codeErr.Cause() != nil {
		t.Fatal("new CodeError should not wrap an error")
	}
	if resp := codeErr.ToResponse(); resp.Code != 123 || resp.Msg != "custom" {
		t.Fatalf("ToResponse = %+v", resp)
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		msg  string
	}{
		{name: "server busy", err: NewServerBusy(), code: ServerBusy},
		{name: "invalid param", err: NewInvalidParam("bad"), code: InvalidParam, msg: "bad"},
		{name: "unauthorized", err: NewUnauthorized("login"), code: Unauthorized, msg: "login"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var codeErr *CodeError
			if !errors.As(tt.err, &codeErr) {
				t.Fatalf("error type = %T, want *CodeError", tt.err)
			}
			if codeErr.Code != tt.code {
				t.Fatalf("code = %d, want %d", codeErr.Code, tt.code)
			}
			if tt.msg != "" && codeErr.Msg != tt.msg {
				t.Fatalf("msg = %q, want %q", codeErr.Msg, tt.msg)
			}
		})
	}
}

func TestWrapAndCause(t *testing.T) {
	base := errors.New("base")

	wrappedBusy := WrapServerBusy(base)
	var codeErr *CodeError
	if !errors.As(wrappedBusy, &codeErr) {
		t.Fatalf("WrapServerBusy type = %T, want *CodeError", wrappedBusy)
	}
	if !errors.Is(wrappedBusy, base) {
		t.Fatal("WrapServerBusy should unwrap base error")
	}
	if !strings.Contains(wrappedBusy.Error(), "base") {
		t.Fatalf("wrapped Error() = %q, want base message", wrappedBusy.Error())
	}

	if got := WithMessage(base, "context").Error(); !strings.Contains(got, "context") {
		t.Fatalf("WithMessage = %q, want context", got)
	}
	if got := WithMessagef(base, "context %s", "x").Error(); !strings.Contains(got, "context x") {
		t.Fatalf("WithMessagef = %q, want formatted context", got)
	}
	if got := Wrap(base, "wrap").Error(); !strings.Contains(got, "wrap") {
		t.Fatalf("Wrap = %q, want wrap", got)
	}
	if got := Wrapf(base, "wrap %s", "x").Error(); !strings.Contains(got, "wrap x") {
		t.Fatalf("Wrapf = %q, want formatted wrap", got)
	}
	if Cause(Wrap(base, "wrap")) != base {
		t.Fatal("Cause should return base error")
	}
}

func TestHandleDaoError(t *testing.T) {
	codeErr := NewInvalidParam("bad")
	if got := HandleDaoError(codeErr, "query"); got != codeErr {
		t.Fatal("HandleDaoError should return CodeError unchanged")
	}

	base := errors.New("db down")
	got := HandleDaoError(base, "query user")
	var wrapped *CodeError
	if !errors.As(got, &wrapped) {
		t.Fatalf("HandleDaoError type = %T, want *CodeError", got)
	}
	if wrapped.Code != ServerBusy {
		t.Fatalf("code = %d, want %d", wrapped.Code, ServerBusy)
	}
	if !strings.Contains(wrapped.err.Error(), "query user") {
		t.Fatalf("wrapped err = %v, want context", wrapped.err)
	}
}

func TestHandleResponseAndHandleError(t *testing.T) {
	serverBusyMsg := NewServerBusy().(*CodeError).Msg
	tests := []struct {
		name string
		err  *CodeError
		code int
		msg  string
	}{
		{name: "server busy hides detail", err: WrapServerBusy(errors.New("db")).(*CodeError), code: ServerBusy, msg: serverBusyMsg},
		{name: "invalid param", err: NewInvalidParam("bad").(*CodeError), code: InvalidParam, msg: "bad"},
		{name: "unauthorized", err: NewUnauthorized("login").(*CodeError), code: Unauthorized, msg: "login"},
		{name: "custom", err: &CodeError{Code: 999, Msg: "custom", err: errors.New("x")}, code: 999, msg: "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.err.HandleResponse()
			if resp.Code != tt.code || resp.Msg != tt.msg {
				t.Fatalf("HandleResponse = %+v, want code=%d msg=%q", resp, tt.code, tt.msg)
			}

			resp = HandleError(tt.err)
			if resp.Code != tt.code || resp.Msg != tt.msg {
				t.Fatalf("HandleError = %+v, want code=%d msg=%q", resp, tt.code, tt.msg)
			}
		})
	}

	resp := HandleError(errors.New("unknown"))
	if resp.Code != ServerBusy || resp.Msg != serverBusyMsg {
		t.Fatalf("unknown HandleError = %+v", resp)
	}
}
