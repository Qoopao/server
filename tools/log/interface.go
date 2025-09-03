package log

import (
	"context"
)

// Logger 日志接口
type Logger interface {
	// Debug 调试日志
	Debug(msg string, fields ...Field)
	DebugContext(ctx context.Context, msg string, fields ...Field)

	// Info 信息日志
	Info(msg string, fields ...Field)
	InfoContext(ctx context.Context, msg string, fields ...Field)

	// Warn 警告日志
	Warn(msg string, fields ...Field)
	WarnContext(ctx context.Context, msg string, fields ...Field)

	// Error 错误日志
	Error(msg string, fields ...Field)
	ErrorContext(ctx context.Context, msg string, fields ...Field)

	// Fatal 致命错误日志（会退出程序）
	Fatal(msg string, fields ...Field)
	FatalContext(ctx context.Context, msg string, fields ...Field)

	// Panic 恐慌日志（会触发 panic）
	Panic(msg string, fields ...Field)
	PanicContext(ctx context.Context, msg string, fields ...Field)

	// With 创建带有字段的子日志器
	With(fields ...Field) Logger

	// Sync 同步日志缓冲区
	Sync() error
}

// Field 日志字段接口
type Field interface {
	// Key 返回字段的键名
	Key() string
	// Value 返回字段的值
	Value() interface{}
}

// Config 日志配置
type Config struct {
	// Level 日志级别: debug, info, warn, error, fatal, panic
	Level string `json:"level" yaml:"level"`

	// Format 日志格式: json, console
	Format string `json:"format" yaml:"format"`

	// OutputPaths 输出路径列表
	OutputPaths []string `json:"output_paths" yaml:"output_paths"`

	// ErrorOutputPaths 错误输出路径列表
	ErrorOutputPaths []string `json:"error_output_paths" yaml:"error_output_paths"`

	// Development 是否为开发模式
	Development bool `json:"development" yaml:"development"`

	// DisableCaller 是否禁用调用者信息
	DisableCaller bool `json:"disable_caller" yaml:"disable_caller"`

	// DisableStacktrace 是否禁用堆栈跟踪
	DisableStacktrace bool `json:"disable_stacktrace" yaml:"disable_stacktrace"`

	// Sampling 采样配置
	Sampling *SamplingConfig `json:"sampling" yaml:"sampling"`

	// EncoderConfig 编码器配置
	EncoderConfig *EncoderConfig `json:"encoder_config" yaml:"encoder_config"`
}

// SamplingConfig 采样配置
type SamplingConfig struct {
	Initial    int `json:"initial" yaml:"initial"`
	Thereafter int `json:"thereafter" yaml:"thereafter"`
}

// EncoderConfig 编码器配置
type EncoderConfig struct {
	// TimeKey 时间字段的键名
	TimeKey string `json:"time_key" yaml:"time_key"`

	// LevelKey 级别字段的键名
	LevelKey string `json:"level_key" yaml:"level_key"`

	// NameKey 名称字段的键名
	NameKey string `json:"name_key" yaml:"name_key"`

	// CallerKey 调用者字段的键名
	CallerKey string `json:"caller_key" yaml:"caller_key"`

	// MessageKey 消息字段的键名
	MessageKey string `json:"message_key" yaml:"message_key"`

	// StacktraceKey 堆栈跟踪字段的键名
	StacktraceKey string `json:"stacktrace_key" yaml:"stacktrace_key"`

	// LineEnding 行结束符
	LineEnding string `json:"line_ending" yaml:"line_ending"`

	// EncodeLevel 级别编码方式
	EncodeLevel string `json:"encode_level" yaml:"encode_level"`

	// EncodeTime 时间编码方式
	EncodeTime string `json:"encode_time" yaml:"encode_time"`

	// EncodeDuration 持续时间编码方式
	EncodeDuration string `json:"encode_duration" yaml:"encode_duration"`

	// EncodeCaller 调用者编码方式
	EncodeCaller string `json:"encode_caller" yaml:"encode_caller"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:             "info",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling: &SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: &EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     "\n",
			EncodeLevel:    "lowercase",
			EncodeTime:     "iso8601",
			EncodeDuration: "string",
			EncodeCaller:   "short",
		},
	}
}

// DevelopmentConfig 返回开发环境配置
func DevelopmentConfig() *Config {
	config := DefaultConfig()
	config.Level = "debug"
	config.Format = "console"
	config.Development = true
	return config
}

// ProductionConfig 返回生产环境配置
func ProductionConfig() *Config {
	config := DefaultConfig()
	config.Level = "info"
	config.Format = "json"
	config.Development = false
	config.DisableCaller = true
	config.DisableStacktrace = true
	return config
}
