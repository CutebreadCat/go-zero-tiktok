package svc

const (
	aiLimitSeconds     = 60
	aiLimitMaxRequests = 20
	aiLimitKeyPrefix   = "ai_limit"

	wsLimitSeconds     = 1
	wsLimitMaxRequests = 30
	wsLimitKeyPrefix   = "ws_limit"

	httpLimitSeconds     = 1
	httpLimitMaxRequests = 100
	httpLimitKeyPrefix   = "http_limit"
)
