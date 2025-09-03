package log

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger zap 日志实现
type zapLogger struct {
	logger *zap.Logger
}

// NewZapLogger 创建新的 zap 日志器
func NewZapLogger(config *Config) (Logger, error) {
	zapConfig, err := buildZapConfig(config)
	if err != nil {
		return nil, fmt.Errorf("构建 zap 配置失败: %w", err)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("构建 zap 日志器失败: %w", err)
	}

	return &zapLogger{logger: logger}, nil
}

// NewDevelopmentLogger 创建开发环境日志器
func NewDevelopmentLogger() (Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("创建开发环境日志器失败: %w", err)
	}
	return &zapLogger{logger: logger}, nil
}

// NewProductionLogger 创建生产环境日志器
func NewProductionLogger() (Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("创建生产环境日志器失败: %w", err)
	}
	return &zapLogger{logger: logger}, nil
}

// buildZapConfig 构建 zap 配置
func buildZapConfig(config *Config) (*zap.Config, error) {
	zapConfig := &zap.Config{
		Level:             parseLevel(config.Level),
		Development:       config.Development,
		DisableCaller:     config.DisableCaller,
		DisableStacktrace: config.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    config.Sampling.Initial,
			Thereafter: config.Sampling.Thereafter,
		},
		Encoding:         config.Format,
		EncoderConfig:    buildEncoderConfig(config.EncoderConfig),
		OutputPaths:      config.OutputPaths,
		ErrorOutputPaths: config.ErrorOutputPaths,
	}

	return zapConfig, nil
}

// buildEncoderConfig 构建编码器配置
func buildEncoderConfig(encoderConfig *EncoderConfig) zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        encoderConfig.TimeKey,
		LevelKey:       encoderConfig.LevelKey,
		NameKey:        encoderConfig.NameKey,
		CallerKey:      encoderConfig.CallerKey,
		MessageKey:     encoderConfig.MessageKey,
		StacktraceKey:  encoderConfig.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    parseEncodeLevel(encoderConfig.EncodeLevel),
		EncodeTime:     parseEncodeTime(encoderConfig.EncodeTime),
		EncodeDuration: parseEncodeDuration(encoderConfig.EncodeDuration),
		EncodeCaller:   parseEncodeCaller(encoderConfig.EncodeCaller),
	}
}

// parseLevel 解析日志级别
func parseLevel(level string) zap.AtomicLevel {
	switch strings.ToLower(level) {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn", "warning":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

// parseEncodeLevel 解析级别编码方式
func parseEncodeLevel(encode string) zapcore.LevelEncoder {
	switch strings.ToLower(encode) {
	case "lowercase":
		return zapcore.LowercaseLevelEncoder
	case "lowercasecolor":
		return zapcore.LowercaseColorLevelEncoder
	case "capital":
		return zapcore.CapitalLevelEncoder
	case "capitalcolor":
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

// parseEncodeTime 解析时间编码方式
func parseEncodeTime(encode string) zapcore.TimeEncoder {
	switch strings.ToLower(encode) {
	case "epoch":
		return zapcore.EpochTimeEncoder
	case "epochmillis":
		return zapcore.EpochMillisTimeEncoder
	case "epochnanos":
		return zapcore.EpochNanosTimeEncoder
	case "iso8601":
		return zapcore.ISO8601TimeEncoder
	case "rfc3339":
		return zapcore.RFC3339TimeEncoder
	case "rfc3339nano":
		return zapcore.RFC3339NanoTimeEncoder
	default:
		return zapcore.ISO8601TimeEncoder
	}
}

// parseEncodeDuration 解析持续时间编码方式
func parseEncodeDuration(encode string) zapcore.DurationEncoder {
	switch strings.ToLower(encode) {
	case "string":
		return zapcore.StringDurationEncoder
	case "nanos":
		return zapcore.NanosDurationEncoder
	case "millis":
		return zapcore.MillisDurationEncoder
	case "seconds":
		return zapcore.SecondsDurationEncoder
	default:
		return zapcore.StringDurationEncoder
	}
}

// parseEncodeCaller 解析调用者编码方式
func parseEncodeCaller(encode string) zapcore.CallerEncoder {
	switch strings.ToLower(encode) {
	case "short":
		return zapcore.ShortCallerEncoder
	case "full":
		return zapcore.FullCallerEncoder
	default:
		return zapcore.ShortCallerEncoder
	}
}

// Debug 调试日志
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, convertFields(fields)...)
}

// DebugContext 带上下文的调试日志
func (l *zapLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {
	l.logger.Debug(msg, convertFields(fields)...)
}

// Info 信息日志
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, convertFields(fields)...)
}

// InfoContext 带上下文的信息日志
func (l *zapLogger) InfoContext(ctx context.Context, msg string, fields ...Field) {
	l.logger.Info(msg, convertFields(fields)...)
}

// Warn 警告日志
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, convertFields(fields)...)
}

// WarnContext 带上下文的警告日志
func (l *zapLogger) WarnContext(ctx context.Context, msg string, fields ...Field) {
	l.logger.Warn(msg, convertFields(fields)...)
}

// Error 错误日志
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, convertFields(fields)...)
}

// ErrorContext 带上下文的错误日志
func (l *zapLogger) ErrorContext(ctx context.Context, msg string, fields ...Field) {
	l.logger.Error(msg, convertFields(fields)...)
}

// Fatal 致命错误日志
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, convertFields(fields)...)
}

// FatalContext 带上下文的致命错误日志
func (l *zapLogger) FatalContext(ctx context.Context, msg string, fields ...Field) {
	l.logger.Fatal(msg, convertFields(fields)...)
}

// Panic 恐慌日志
func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.logger.Panic(msg, convertFields(fields)...)
}

// PanicContext 带上下文的恐慌日志
func (l *zapLogger) PanicContext(ctx context.Context, msg string, fields ...Field) {
	l.logger.Panic(msg, convertFields(fields)...)
}

// With 创建带有字段的子日志器
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(convertFields(fields)...)}
}

// Sync 同步日志缓冲区
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// convertFields 转换字段格式
func convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key(), field.Value())
	}
	return zapFields
}

// 预定义的常用字段类型
type (
	// StringField 字符串字段
	StringField struct {
		key   string
		value string
	}

	// IntField 整数字段
	IntField struct {
		key   string
		value int
	}

	// Int64Field 64位整数字段
	Int64Field struct {
		key   string
		value int64
	}

	// Float64Field 64位浮点数字段
	Float64Field struct {
		key   string
		value float64
	}

	// BoolField 布尔字段
	BoolField struct {
		key   string
		value bool
	}

	// AnyField 任意类型字段
	AnyField struct {
		key   string
		value interface{}
	}

	// ErrorField 错误字段
	ErrorField struct {
		key   string
		value error
	}
)

// 实现 Field 接口
func (f StringField) Key() string        { return f.key }
func (f StringField) Value() interface{} { return f.value }

func (f IntField) Key() string        { return f.key }
func (f IntField) Value() interface{} { return f.value }

func (f Int64Field) Key() string        { return f.key }
func (f Int64Field) Value() interface{} { return f.value }

func (f Float64Field) Key() string        { return f.key }
func (f Float64Field) Value() interface{} { return f.value }

func (f BoolField) Key() string        { return f.key }
func (f BoolField) Value() interface{} { return f.value }

func (f AnyField) Key() string        { return f.key }
func (f AnyField) Value() interface{} { return f.value }

func (f ErrorField) Key() string        { return f.key }
func (f ErrorField) Value() interface{} { return f.value }

// 便捷的字段创建函数
func String(key, value string) Field {
	return StringField{key: key, value: value}
}

func Int(key string, value int) Field {
	return IntField{key: key, value: value}
}

func Int64(key string, value int64) Field {
	return Int64Field{key: key, value: value}
}

func Float64(key string, value float64) Field {
	return Float64Field{key: key, value: value}
}

func Bool(key string, value bool) Field {
	return BoolField{key: key, value: value}
}

func Any(key string, value interface{}) Field {
	return AnyField{key: key, value: value}
}

func Error(err error) Field {
	return ErrorField{key: "error", value: err}
}

func NamedError(key string, err error) Field {
	return ErrorField{key: key, value: err}
}
