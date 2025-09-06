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

func TestJSONFormat(t *testing.T) {
	// 创建一个测试 logger
	logger := NewOtelLogger("test-logger")
	logger.SetLevel(LevelInfo)

	// 确保使用 JSON 格式
	logger.SetJSONFormat(true)

	// 测试基本日志
	logger.Info("Test message")

	// 测试格式化日志
	logger.Infof("Processing %d items", 42)

	// 测试键值对参数（模拟您的使用场景）
	logger.Infof("sendMessages completed",
		"total_messages", 1,
		"success_count", 1)

	// 测试上下文日志
	ctx := context.Background()
	logger.CtxInfof(ctx, "sendMessages completed",
		"total_messages", 5,
		"success_count", 3,
		"error_count", 2)
}

func TestKeyValueParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string
		fields   map[string]interface{}
	}{
		{
			name:     "No args",
			args:     []interface{}{},
			expected: "",
			fields:   nil,
		},
		{
			name:     "Single arg",
			args:     []interface{}{"hello"},
			expected: "hello",
			fields:   nil,
		},
		{
			name:     "Key-value pairs",
			args:     []interface{}{"total_messages", 5, "success_count", 3},
			expected: "",
			fields: map[string]interface{}{
				"total_messages": 5,
				"success_count":  3,
			},
		},
		{
			name:     "Mixed args",
			args:     []interface{}{"Processing", "total_messages", 5},
			expected: "Processing total_messages 5",
			fields:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message, fields := parseKeyValuePairs(test.args)

			if message != test.expected {
				t.Errorf("Expected message '%s', got '%s'", test.expected, message)
			}

			if test.fields == nil && fields != nil {
				t.Errorf("Expected nil fields, got %v", fields)
			} else if test.fields != nil {
				if fields == nil {
					t.Errorf("Expected fields %v, got nil", test.fields)
				} else {
					if len(fields) != len(test.fields) {
						t.Errorf("Expected %d fields, got %d", len(test.fields), len(fields))
					}
					for k, v := range test.fields {
						if fields[k] != v {
							t.Errorf("Expected field %s=%v, got %v", k, v, fields[k])
						}
					}
				}
			}
		})
	}
}

func TestJSONLogCreation(t *testing.T) {
	logger := NewOtelLogger("test-logger")

	// 测试基本 JSON 日志
	jsonLog := logger.createJSONLog("INFO", "Test message", nil)
	t.Logf("Basic JSON log: %s", jsonLog)

	// 测试带字段的 JSON 日志
	fields := map[string]interface{}{
		"total_messages": 5,
		"success_count":  3,
		"error_count":    2,
	}
	jsonLogWithFields := logger.createJSONLog("INFO", "sendMessages completed", fields)
	t.Logf("JSON log with fields: %s", jsonLogWithFields)
}
