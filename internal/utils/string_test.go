package myutils

import (
	"net/http"
	"testing"
)

func TestParseCookieTostring(t *testing.T) {
	got := ParseCookieTostring([]*http.Cookie{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
	})
	if got != "a=1; b=2" {
		t.Fatalf("cookie string = %q, want %q", got, "a=1; b=2")
	}

	if got := ParseCookieTostring(nil); got != "" {
		t.Fatalf("empty cookie string = %q, want empty", got)
	}
}
