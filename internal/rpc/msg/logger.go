package msg

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger 初始化zap日志
func InitLogger() error {
	// 确保logs目录存在
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// 配置日志输出
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{
		filepath.Join(logsDir, "msg.log"), // 消息服务日志
	}
	config.ErrorOutputPaths = []string{
		filepath.Join(logsDir, "error.log"), // 错误日志
	}

	// 设置日志级别
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	// 设置日志格式
	config.Encoding = "json"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// 创建logger
	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	// 设置全局logger
	zap.ReplaceGlobals(Logger)

	return nil
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// GetLogger 获取logger实例
func GetLogger() *zap.Logger {
	return Logger
}
