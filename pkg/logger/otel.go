package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RegisterOTelTraceExtractor 注册 OTel span 提取器:
// 从 context 中读取当前 OTel span,把 TraceID / SpanID 输出为
// trace_id / span_id 日志字段。开启 go-zero 内置 Telemetry 后,
// rest / zrpc 会自动把 span 放入 context,此提取器随之生效。
//
// 未开启 Telemetry 或 context 中无有效 span 时返回空字段,不影响日志输出,
// 因此各服务可以无条件注册。
func RegisterOTelTraceExtractor() {
	AddContextFieldExtractor(func(ctx context.Context) []zap.Field {
		sc := trace.SpanContextFromContext(ctx)
		if !sc.IsValid() {
			return nil
		}
		fields := make([]zap.Field, 0, 2)
		if sc.HasTraceID() {
			fields = append(fields, zap.String("trace_id", sc.TraceID().String()))
		}
		if sc.HasSpanID() {
			fields = append(fields, zap.String("span_id", sc.SpanID().String()))
		}
		return fields
	})
}
