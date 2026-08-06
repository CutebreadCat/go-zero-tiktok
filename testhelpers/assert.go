package testhelpers

import (
	"errors"
	"testing"

	"go_zero-tiktok/pkg/xerr"
)

// AssertNoErr 断言 err 为 nil
func AssertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// AssertErr 断言 err 非 nil
func AssertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertInvalidParam 断言 err 是 xerr 参数错误（code==1002）
func AssertInvalidParam(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected InvalidParam error, got nil")
	}
	var ce *xerr.CodeError
	if !errors.As(err, &ce) {
		t.Fatalf("expected xerr.CodeError, got %T: %v", err, err)
	}
	if ce.Code != xerr.InvalidParam {
		t.Fatalf("expected InvalidParam code %d, got %d (%s)", xerr.InvalidParam, ce.Code, ce.Msg)
	}
}

// AssertUnauthorized 断言 err 是 xerr 鉴权错误（code==401）
func AssertUnauthorized(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected Unauthorized error, got nil")
	}
	var ce *xerr.CodeError
	if !errors.As(err, &ce) {
		t.Fatalf("expected xerr.CodeError, got %T: %v", err, err)
	}
	if ce.Code != xerr.Unauthorized {
		t.Fatalf("expected Unauthorized code %d, got %d (%s)", xerr.Unauthorized, ce.Code, ce.Msg)
	}
}

// AssertEqual 断言两值相等
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("assert equal failed: got %v, want %v", got, want)
	}
}
