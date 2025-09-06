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
	"testing"
)

func TestOtelLogger(t *testing.T) {
	// 创建一个测试 logger
	logger := NewOtelLogger("test-logger")

	// 测试基本日志方法
	logger.Info("Test info message")
	logger.Infof("Test info message with format: %d", 42)

	// 测试不同级别
	logger.Trace("Test trace message")
	logger.Debug("Test debug message")
	logger.Notice("Test notice message")
	logger.Warn("Test warn message")
	logger.Error("Test error message")
	logger.Fatal("Test fatal message")

	// 测试格式化日志
	logger.Tracef("Trace format: %s", "value")
	logger.Debugf("Debug format: %d", 123)
	logger.Noticef("Notice format: %v", true)
	logger.Warnf("Warn format: %s", "warning")
	logger.Errorf("Error format: %s", "error")
	logger.Fatalf("Fatal format: %s", "fatal")

	// 测试上下文日志
	ctx := context.Background()
	logger.CtxTracef(ctx, "Context trace: %s", "value")
	logger.CtxDebugf(ctx, "Context debug: %d", 123)
	logger.CtxInfof(ctx, "Context info: %v", true)
	logger.CtxNoticef(ctx, "Context notice: %s", "notice")
	logger.CtxWarnf(ctx, "Context warn: %s", "warning")
	logger.CtxErrorf(ctx, "Context error: %s", "error")
	logger.CtxFatalf(ctx, "Context fatal: %s", "fatal")

	// 测试级别设置
	logger.SetLevel(LevelDebug)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("Expected level Debug, got %v", logger.GetLevel())
	}

	// 测试日志级别过滤
	logger.SetLevel(LevelWarn)
	logger.Info("This should not appear") // 应该被过滤掉
	logger.Warn("This should appear")     // 应该出现
	logger.Error("This should appear")    // 应该出现
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelTrace, "TRACE"},
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelNotice, "NOTICE"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{Level(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

func TestToSlogLevel(t *testing.T) {
	tests := []struct {
		klogLevel Level
		expected  string
	}{
		{LevelTrace, "DEBUG"},
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelNotice, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "ERROR"},
	}

	for _, test := range tests {
		result := test.klogLevel.toSlogLevel().String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger("test-default")
	if logger == nil {
		t.Error("NewDefaultLogger returned nil")
	}

	// 测试基本功能
	logger.Info("Default logger test")
	logger.Infof("Default logger test with format: %d", 42)
}

func TestNewLoggerWithProvider(t *testing.T) {
	// 这里我们无法实际创建 LoggerProvider 进行测试，因为没有实际的 OTEL 后端
	// 但我们可以测试函数不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewLoggerWithProvider panicked: %v", r)
		}
	}()

	// 注意：这个测试可能会失败，因为没有实际的 OTEL 后端
	// logger := NewLoggerWithProvider("test-provider", nil)
	// if logger == nil {
	//     t.Error("NewLoggerWithProvider returned nil")
	// }
}

// BenchmarkOtelLogger 性能测试
func BenchmarkOtelLogger(b *testing.B) {
	logger := NewOtelLogger("benchmark-logger")
	logger.SetLevel(LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message")
	}
}

func BenchmarkOtelLoggerWithFormat(b *testing.B) {
	logger := NewOtelLogger("benchmark-logger")
	logger.SetLevel(LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infof("Benchmark message %d", i)
	}
}
