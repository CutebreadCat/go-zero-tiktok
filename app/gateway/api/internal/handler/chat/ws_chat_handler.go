// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import "net/http"

func WsChatHandler(_ interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "websocket gateway migration is pending", http.StatusNotImplemented)
	}
}
