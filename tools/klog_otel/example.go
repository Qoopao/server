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
	"time"
)

// ExampleUsage 展示如何使用 klog_otel
func ExampleUsage() {
	ctx := context.Background()

	// 方法1: 使用完整的设置流程
	logger, lp, err := SetupLogger(ctx, "example-service")
	if err != nil {
		panic(err)
	}
	defer lp.Shutdown(ctx)

	// 设置日志级别
	logger.SetLevel(LevelInfo)

	// 使用各种日志方法
	logger.Info("Service starting up")
	logger.Infof("Processing request %d", 123)

	// 使用上下文日志
	type requestIDKey string
	reqCtx := context.WithValue(ctx, requestIDKey("request_id"), "req-456")
	logger.CtxInfof(reqCtx, "Handling request %s", "GET /api/users")

	// 错误日志
	logger.Errorf("Failed to connect to database: %s", "connection timeout")

	// 方法2: 使用默认 logger（假设全局 LoggerProvider 已经设置）
	defaultLogger := NewDefaultLogger("another-service")
	defaultLogger.Warn("This is a warning message")

	// 方法3: 使用配置参数直接创建
	customLogger, lp2, err := SetupLoggerWithConfig(ctx, "custom-service", "localhost:4317", true)
	if err != nil {
		panic(err)
	}
	defer lp2.Shutdown(ctx)
	customLogger.Info("Using custom provider with config")

	// 方法4: 使用不安全的端点
	insecureLogger, lp3, err := SetupLoggerWithInsecureEndpoint(ctx, "insecure-service", "localhost:4317")
	if err != nil {
		panic(err)
	}
	defer lp3.Shutdown(ctx)
	insecureLogger.Info("Using insecure endpoint")
}

// ExampleWithKitex 展示如何在 Kitex 中使用
func ExampleWithKitex() {
	ctx := context.Background()

	// 初始化 OTEL log provider
	logger, lp, err := SetupLogger(ctx, "kitex-service")
	if err != nil {
		panic("failed to initialize logger")
	}
	defer lp.Shutdown(ctx)

	// 设置日志级别
	logger.SetLevel(LevelInfo)

	// 现在可以使用 logger 作为 klog 的实现
	// 例如：klog.SetLogger(logger) // 如果 klog 支持设置自定义 logger

	// 使用示例
	logger.Info("Kitex service started successfully")
	logger.Infof("Listening on port %d", 8080)

	// 模拟请求处理
	type traceIDKey string
	reqCtx := context.WithValue(ctx, traceIDKey("trace_id"), "trace-123")
	logger.CtxInfof(reqCtx, "Processing request")

	// 模拟错误
	logger.Errorf("Database connection failed after %d retries", 3)
}

// ExampleLevels 展示不同日志级别的使用
func ExampleLevels() {
	ctx := context.Background()

	logger, lp, err := SetupLogger(ctx, "level-demo")
	if err != nil {
		panic(err)
	}
	defer lp.Shutdown(ctx)

	// 设置为 Debug 级别，可以看到所有日志
	logger.SetLevel(LevelDebug)

	logger.Trace("This is trace level")
	logger.Debug("This is debug level")
	logger.Info("This is info level")
	logger.Notice("This is notice level")
	logger.Warn("This is warn level")
	logger.Error("This is error level")
	logger.Fatal("This is fatal level")

	// 格式化日志
	logger.Tracef("Trace with format: %s", "value")
	logger.Debugf("Debug with format: %d", 42)
	logger.Infof("Info with format: %v", time.Now())
	logger.Noticef("Notice with format: %s", "important")
	logger.Warnf("Warn with format: %s", "warning")
	logger.Errorf("Error with format: %s", "error")
	logger.Fatalf("Fatal with format: %s", "fatal")

	// 上下文日志
	logger.CtxTracef(ctx, "Context trace: %s", "value")
	logger.CtxDebugf(ctx, "Context debug: %d", 42)
	logger.CtxInfof(ctx, "Context info: %v", time.Now())
	logger.CtxNoticef(ctx, "Context notice: %s", "important")
	logger.CtxWarnf(ctx, "Context warn: %s", "warning")
	logger.CtxErrorf(ctx, "Context error: %s", "error")
	logger.CtxFatalf(ctx, "Context fatal: %s", "fatal")
}
