package middleware

const (
	DefaultRateLimitSeconds     = 1
	DefaultRateLimitMaxRequests = 100
	DefaultRateLimitKeyPrefix   = "http_limit"

	rateLimitRemoteAddrKey = "remote_addr"
)
