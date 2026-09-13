package logger

const (
	// LogFilePath 是日志根目录下的叶子目录名，最终路径为
	// {项目根}/{LogFilePath}/{service}/{date}/xxx.log，与进程 cwd 无关。
	LogFilePath = "logs"

	// LogDirEnv 用于覆盖日志根目录（{service}/{date}/ 的父目录）。
	// 未设置时自动定位项目根下的 logs/。
	LogDirEnv = "TIKTOK_LOG_DIR"

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
