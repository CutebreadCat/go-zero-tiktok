package logger

const (
	// LogFilePath 对应 ${pwd}/{LogFilePath}/{service}/{date}/xxx.log 相对于当前运行路径而言
	LogFilePath = "logs"

	LogFilePathTemplate      = "%s/%s/%s/%s.log"
	ErrorLogFilePathTemplate = "%s/%s/%s/%s_stderr.log"

	// DefaultLogLevel 是默认的日志等级. Supported Level: debug info warn error fatal
	DefaultLogLevel = "INFO"

	StackTraceKey = "stacktrace"
	ServiceKey    = "service"
	SourceKey     = "source"

	KlogSource  = "klog"
	MysqlSource = "mysql"
	RedisSource = "redis"
	KafkaSource = "kafka"
)
