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

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// InitLog 初始化 OpenTelemetry LoggerProvider
// 这个函数参考了您提供的 initLog 函数
func InitLog(ctx context.Context) (*sdklog.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewSimpleProcessor(logExporter),
		),
	)

	global.SetLoggerProvider(lp)
	return lp, nil
}

// InitLogWithEndpoint 使用指定的端点初始化 OpenTelemetry LoggerProvider
func InitLogWithEndpoint(ctx context.Context, endpoint string) (*sdklog.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log exporter with endpoint %s: %w", endpoint, err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewSimpleProcessor(logExporter),
		),
	)

	global.SetLoggerProvider(lp)
	return lp, nil
}

// NewDefaultLogger 创建一个使用全局 LoggerProvider 的默认 logger
func NewDefaultLogger(loggerName string) FullLogger {
	return NewOtelLogger(loggerName)
}

// NewLoggerWithProvider 使用指定的 LoggerProvider 创建 logger
func NewLoggerWithProvider(loggerName string, provider *sdklog.LoggerProvider) FullLogger {
	return NewOtelLoggerWithProvider(loggerName, provider)
}

// NewKitexCompatibleLogger 创建与 Kitex 兼容的 logger
func NewKitexCompatibleLogger(loggerName string) klog.FullLogger {
	return NewKitexLogger(loggerName)
}

// SetupLogger 完整的设置流程，包括初始化 LoggerProvider 和创建 logger
func SetupLogger(ctx context.Context, loggerName string) (FullLogger, *sdklog.LoggerProvider, error) {
	lp, err := InitLog(ctx)
	if err != nil {
		return nil, nil, err
	}

	logger := NewOtelLogger(loggerName)
	return logger, lp, nil
}

// SetupLoggerWithEndpoint 使用指定端点设置 logger
func SetupLoggerWithEndpoint(ctx context.Context, loggerName, endpoint string) (FullLogger, *sdklog.LoggerProvider, error) {
	lp, err := InitLogWithEndpoint(ctx, endpoint)
	if err != nil {
		return nil, nil, err
	}

	logger := NewOtelLogger(loggerName)
	return logger, lp, nil
}

// SetupLoggerWithConfig 使用配置参数设置 logger
func SetupLoggerWithConfig(ctx context.Context, loggerName, endpoint string, insecure bool) (FullLogger, *sdklog.LoggerProvider, error) {
	return NewOtelLoggerWithConfig(ctx, loggerName, endpoint, insecure)
}

// SetupLoggerWithInsecureEndpoint 使用不安全的端点设置 logger（默认 localhost:4317）
func SetupLoggerWithInsecureEndpoint(ctx context.Context, loggerName, endpoint string) (FullLogger, *sdklog.LoggerProvider, error) {
	return NewOtelLoggerWithConfig(ctx, loggerName, endpoint, true)
}
