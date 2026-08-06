package myutils

import (
	"net/http"
	"strings"
)

func ParseCookieTostring(cookie []*http.Cookie) string {
	var cookieStr []string
	for _, c := range cookie {
		cookieStr = append(cookieStr, c.Name+"="+c.Value)
	}
	return strings.Join(cookieStr, "; ")
}
