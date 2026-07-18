package ctxkey

type Key string

const (
	UserID  Key = "user_id"
	TraceID Key = "trace_id"
	SpanID  Key = "span_id"
)
