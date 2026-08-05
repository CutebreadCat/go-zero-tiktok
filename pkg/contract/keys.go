package types

// ContextKey context 键类型，用于在请求链路中存放跨层共享的元数据
// （网关鉴权中间件写入，RPC / logger / utils 读取）。
type ContextKey string

const (
	ContextKeyUserID  ContextKey = "user_id"
	ContextKeyTraceID ContextKey = "trace_id"
	ContextKeySpanID  ContextKey = "span_id"
)
