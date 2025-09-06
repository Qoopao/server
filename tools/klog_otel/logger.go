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
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// OtelLogger 实现 klog 的所有接口，使用 OpenTelemetry 作为后端
type OtelLogger struct {
	logger *slog.Logger
	level  Level
	output io.Writer
	mu     sync.RWMutex
}

// NewOtelLogger 创建新的 OpenTelemetry logger 实例
func NewOtelLogger(loggerName string) *OtelLogger {
	// 使用全局的 LoggerProvider
	otelLogger := otelslog.NewLogger(loggerName)

	return &OtelLogger{
		logger: otelLogger,
		level:  LevelInfo, // 默认级别
		output: nil,       // 不使用本地输出，直接发送到 OTEL
	}
}

// NewOtelLoggerWithProvider 使用指定的 LoggerProvider 创建 logger
func NewOtelLoggerWithProvider(loggerName string, provider *sdklog.LoggerProvider) *OtelLogger {
	// 临时设置全局 provider
	global.SetLoggerProvider(provider)

	otelLogger := otelslog.NewLogger(loggerName)

	return &OtelLogger{
		logger: otelLogger,
		level:  LevelInfo,
		output: nil,
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
		logger: otelLogger,
		level:  LevelInfo,
		output: nil,
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

// Logger 接口实现
func (ol *OtelLogger) Trace(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelTrace {
		ol.logger.Debug(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Debug(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelDebug {
		ol.logger.Debug(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Info(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelInfo {
		ol.logger.Info(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Notice(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelNotice {
		ol.logger.Info(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Warn(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelWarn {
		ol.logger.Warn(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Error(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelError {
		ol.logger.Error(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Fatal(v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelFatal {
		ol.logger.Error(fmt.Sprint(v...))
	}
	ol.mu.RUnlock()
}

// FormatLogger 接口实现
func (ol *OtelLogger) Tracef(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelTrace {
		ol.logger.Debug(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Debugf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelDebug {
		ol.logger.Debug(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Infof(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelInfo {
		ol.logger.Info(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Noticef(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelNotice {
		ol.logger.Info(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Warnf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelWarn {
		ol.logger.Warn(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Errorf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelError {
		ol.logger.Error(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) Fatalf(format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelFatal {
		ol.logger.Error(fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

// CtxLogger 接口实现
func (ol *OtelLogger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelTrace {
		ol.logger.DebugContext(ctx, fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelDebug {
		ol.logger.DebugContext(ctx, fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelInfo {
		ol.logger.InfoContext(ctx, fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelNotice {
		ol.logger.InfoContext(ctx, fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelWarn {
		ol.logger.WarnContext(ctx, fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelError {
		ol.logger.ErrorContext(ctx, fmt.Sprintf(format, v...))
	}
	ol.mu.RUnlock()
}

func (ol *OtelLogger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	ol.mu.RLock()
	if ol.level <= LevelFatal {
		ol.logger.ErrorContext(ctx, fmt.Sprintf(format, v...))
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
