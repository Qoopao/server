/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package klog_otel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// LogEntry 表示结构化的日志条目
type LogEntry struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Time    string                 `json:"timestamp"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// OtelLogger 实现 klog 的所有接口，使用 OpenTelemetry 作为后端
type OtelLogger struct {
	logger     *slog.Logger
	level      Level
	output     io.Writer
	jsonFormat bool
	mu         sync.RWMutex
}

// NewOtelLogger 创建新的 OpenTelemetry logger 实例
func NewOtelLogger(loggerName string) *OtelLogger {
	// 使用全局的 LoggerProvider
	otelLogger := otelslog.NewLogger(loggerName)

	return &OtelLogger{
		logger:     otelLogger,
		level:      LevelInfo, // 默认级别
		output:     nil,       // 不使用本地输出，直接发送到 OTEL
		jsonFormat: true,      // 默认使用 JSON 格式
	}
}

// NewOtelLoggerWithFormat 创建指定格式的 OpenTelemetry logger 实例
func NewOtelLoggerWithFormat(loggerName string, jsonFormat bool) *OtelLogger {
	// 使用全局的 LoggerProvider
	otelLogger := otelslog.NewLogger(loggerName)

	return &OtelLogger{
		logger:     otelLogger,
		level:      LevelInfo, // 默认级别
		output:     nil,       // 不使用本地输出，直接发送到 OTEL
		jsonFormat: jsonFormat,
	}
}

// NewOtelLoggerWithProvider 使用指定的 LoggerProvider 创建 logger
func NewOtelLoggerWithProvider(loggerName string, provider *sdklog.LoggerProvider) *OtelLogger {
	// 临时设置全局 provider
	global.SetLoggerProvider(provider)

	otelLogger := otelslog.NewLogger(loggerName)

	return &OtelLogger{
		logger:     otelLogger,
		level:      LevelInfo,
		output:     nil,
		jsonFormat: true, // 默认使用 JSON 格式
	}
}

// NewOtelLoggerWithConfig 使用配置参数构建 LoggerProvider 并创建 logger
func NewOtelLoggerWithConfig(ctx context.Context, loggerName string, endpoint string, insecure bool) (*OtelLogger, *sdklog.LoggerProvider, error) {
	// 构建 log exporter
	var logExporter sdklog.Exporter
	var err error

	if insecure {
		logExporter, err = otlploggrpc.New(ctx,
			otlploggrpc.WithInsecure(),
			otlploggrpc.WithEndpoint(endpoint),
		)
	} else {
		logExporter, err = otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(endpoint),
		)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize log exporter: %w", err)
	}

	// 构建 LoggerProvider
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewSimpleProcessor(logExporter),
		),
	)

	// 设置全局 provider
	global.SetLoggerProvider(lp)

	// 创建 logger
	otelLogger := otelslog.NewLogger(loggerName)

	otelLoggerImpl := &OtelLogger{
		logger:     otelLogger,
		level:      LevelInfo,
		output:     nil,
		jsonFormat: true, // 默认使用 JSON 格式
	}

	return otelLoggerImpl, lp, nil
}

// String 返回级别的字符串表示
func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelNotice:
		return "NOTICE"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// toSlogLevel 将 klog 级别转换为 slog 级别
func (l Level) toSlogLevel() slog.Level {
	switch l {
	case LevelTrace, LevelDebug:
		return slog.LevelDebug
	case LevelInfo, LevelNotice:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError // Fatal 在 slog 中也是 Error 级别
	default:
		return slog.LevelInfo
	}
}

// createJSONLog 创建 JSON 格式的日志条目
func (ol *OtelLogger) createJSONLog(level string, message string, fields map[string]interface{}) string {
	entry := LogEntry{
		Level:   level,
		Message: message,
		Time:    fmt.Sprintf("%d", time.Now().UnixNano()/1e6), // 毫秒时间戳
		Fields:  fields,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// 如果 JSON 序列化失败，返回简单的字符串格式
		return fmt.Sprintf(`{"level":"%s","message":"%s","timestamp":"%s","error":"json_marshal_failed"}`, level, message, entry.Time)
	}

	return string(jsonBytes)
}

// logWithFormat 根据配置决定是否使用 JSON 格式输出
func (ol *OtelLogger) logWithFormat(level Level, message string, fields map[string]interface{}) {
	ol.mu.RLock()
	useJSON := ol.jsonFormat
	ol.mu.RUnlock()

	if useJSON {
		jsonLog := ol.createJSONLog(level.String(), message, fields)
		// 使用 slog 输出 JSON 格式的日志
		switch level {
		case LevelTrace, LevelDebug:
			ol.logger.Debug(jsonLog)
		case LevelInfo, LevelNotice:
			ol.logger.Info(jsonLog)
		case LevelWarn:
			ol.logger.Warn(jsonLog)
		case LevelError, LevelFatal:
			ol.logger.Error(jsonLog)
		}
	} else {
		// 使用原始格式
		switch level {
		case LevelTrace, LevelDebug:
			ol.logger.Debug(message)
		case LevelInfo, LevelNotice:
			ol.logger.Info(message)
		case LevelWarn:
			ol.logger.Warn(message)
		case LevelError, LevelFatal:
			ol.logger.Error(message)
		}
	}
}

// Logger 接口实现
func (ol *OtelLogger) Trace(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelTrace {
		ol.logWithFormat(LevelTrace, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Debug(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelDebug {
		ol.logWithFormat(LevelDebug, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Info(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelInfo {
		ol.logWithFormat(LevelInfo, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Notice(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelNotice {
		ol.logWithFormat(LevelNotice, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Warn(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelWarn {
		ol.logWithFormat(LevelWarn, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Error(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelError {
		ol.logWithFormat(LevelError, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Fatal(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelFatal {
		ol.logWithFormat(LevelFatal, fmt.Sprint(v...), nil)
	}
	ol.mu.RUnlock()
}

// FormatLogger 接口实现
func (ol *OtelLogger) Tracef(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelTrace {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelTrace, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Debugf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelDebug {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelDebug, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Infof(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelInfo {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelInfo, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Noticef(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelNotice {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelNotice, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Warnf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelWarn {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelWarn, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Errorf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelError {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelError, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Fatalf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelFatal {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormat(LevelFatal, message, fields)
	}
	ol.mu.RUnlock()
}

// logWithFormatContext 根据配置决定是否使用 JSON 格式输出（带上下文）
func (ol *OtelLogger) logWithFormatContext(ctx context.Context, level Level, message string, fields map[string]interface{}) {
	ol.mu.RLock()
	useJSON := ol.jsonFormat
	ol.mu.RUnlock()

	if useJSON {
		jsonLog := ol.createJSONLog(level.String(), message, fields)
		// 使用 slog 输出 JSON 格式的日志
		switch level {
		case LevelTrace, LevelDebug:
			ol.logger.DebugContext(ctx, jsonLog)
		case LevelInfo, LevelNotice:
			ol.logger.InfoContext(ctx, jsonLog)
		case LevelWarn:
			ol.logger.WarnContext(ctx, jsonLog)
		case LevelError, LevelFatal:
			ol.logger.ErrorContext(ctx, jsonLog)
		}
	} else {
		// 使用原始格式
		switch level {
		case LevelTrace, LevelDebug:
			ol.logger.DebugContext(ctx, message)
		case LevelInfo, LevelNotice:
			ol.logger.InfoContext(ctx, message)
		case LevelWarn:
			ol.logger.WarnContext(ctx, message)
		case LevelError, LevelFatal:
			ol.logger.ErrorContext(ctx, message)
		}
	}
}

// parseKeyValuePairs 解析键值对参数
func parseKeyValuePairs(args []interface{}) (string, map[string]interface{}) {
	if len(args) == 0 {
		return "", nil
	}

	// 检查是否有键值对（偶数个参数且第一个是字符串）
	if len(args)%2 == 0 && len(args) > 0 {
		fields := make(map[string]interface{})
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				if key, ok := args[i].(string); ok {
					fields[key] = args[i+1]
				}
			}
		}
		if len(fields) > 0 {
			return "", fields
		}
	}

	// 否则作为格式化参数处理
	return fmt.Sprint(args...), nil
}

// CtxLogger 接口实现
func (ol *OtelLogger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelTrace {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormatContext(ctx, LevelTrace, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelDebug {
		ol.logger.DebugContext(ctx, format, v...)
		// message, fields := parseKeyValuePairs(v)
		// if message == "" {
		// 	message = format
		// }
		// ol.logWithFormatContext(ctx, LevelDebug, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelInfo {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormatContext(ctx, LevelInfo, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelNotice {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormatContext(ctx, LevelNotice, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelWarn {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormatContext(ctx, LevelWarn, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelError {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormatContext(ctx, LevelError, message, fields)
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelFatal {
		message, fields := parseKeyValuePairs(v)
		if message == "" {
			message = format
		}
		ol.logWithFormatContext(ctx, LevelFatal, message, fields)
	}
	ol.mu.RUnlock()
}

// SetLevel 实现 Control 接口
func (ol *OtelLogger) SetLevel(level Level) {
	ol.mu.Lock()
	ol.level = level
	ol.mu.Unlock()
}

// SetKlogLevel 实现 klog.FullLogger 接口
func (ol *OtelLogger) SetKlogLevel(level klog.Level) {
	ol.mu.Lock()
	ol.level = Level(level)
	ol.mu.Unlock()
}

// SetJSONFormat 设置是否使用 JSON 格式输出
func (ol *OtelLogger) SetJSONFormat(useJSON bool) {
	ol.mu.Lock()
	ol.jsonFormat = useJSON
	ol.mu.Unlock()
}

// IsJSONFormat 返回是否使用 JSON 格式
func (ol *OtelLogger) IsJSONFormat() bool {
	ol.mu.RLock()
	defer ol.mu.RUnlock()
	return ol.jsonFormat
}

func (ol *OtelLogger) SetOutput(w io.Writer) {
	ol.mu.Lock()
	ol.output = w
	ol.mu.Unlock()
}

// GetLevel 获取当前日志级别
func (ol *OtelLogger) GetLevel() Level {
	ol.mu.RLock()
	defer ol.mu.RUnlock()
	return ol.level
}

// GetLogger 获取底层的 slog.Logger
func (ol *OtelLogger) GetLogger() *slog.Logger {
	return ol.logger
}

// KitexLogger 适配器，实现 klog.FullLogger 接口
type KitexLogger struct {
	*OtelLogger
}

// SetLevel 实现 klog.FullLogger 接口
func (kl *KitexLogger) SetLevel(level klog.Level) {
	kl.OtelLogger.SetKlogLevel(level)
}

// SetJSONFormat 设置是否使用 JSON 格式输出
func (kl *KitexLogger) SetJSONFormat(useJSON bool) {
	kl.OtelLogger.SetJSONFormat(useJSON)
}

// IsJSONFormat 返回是否使用 JSON 格式
func (kl *KitexLogger) IsJSONFormat() bool {
	return kl.OtelLogger.IsJSONFormat()
}

// NewKitexLogger 创建适配 klog.FullLogger 接口的 logger
func NewKitexLogger(loggerName string) klog.FullLogger {
	return &KitexLogger{
		OtelLogger: NewOtelLogger(loggerName),
	}
}

// NewKitexLoggerWithProvider 创建适配 klog.FullLogger 接口的 logger
func NewKitexLoggerWithProvider(loggerName string, provider *sdklog.LoggerProvider) klog.FullLogger {
	return &KitexLogger{
		OtelLogger: NewOtelLoggerWithProvider(loggerName, provider),
	}
}

// NewKitexLoggerWithConfig 创建适配 klog.FullLogger 接口的 logger
func NewKitexLoggerWithConfig(ctx context.Context, loggerName, endpoint string, insecure bool) (klog.FullLogger, *sdklog.LoggerProvider, error) {
	otelLogger, lp, err := NewOtelLoggerWithConfig(ctx, loggerName, endpoint, insecure)
	if err != nil {
		return nil, nil, err
	}
	return &KitexLogger{OtelLogger: otelLogger}, lp, nil
}

// 确保 KitexLogger 实现 klog.FullLogger 接口
var _ klog.FullLogger = (*KitexLogger)(nil)
