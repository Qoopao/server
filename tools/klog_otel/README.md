# klog_otel

OpenTelemetry 日志实现，实现了 CloudWeGo Kitex 的 klog 接口。

## 功能特性

- 完整实现 klog 的所有接口：`Logger`, `FormatLogger`, `CtxLogger`, `Control`, `FullLogger`
- 基于 OpenTelemetry 的 `sdklog.LoggerProvider` 实现
- 支持所有日志级别：Trace, Debug, Info, Notice, Warn, Error, Fatal
- 支持上下文日志记录
- 线程安全
- 支持自定义 LoggerProvider
- **完全兼容 CloudWeGo Kitex 的 klog 接口**
- 提供适配器模式，无缝集成到现有 Kitex 项目中
- **自动 JSON 格式输出**，支持结构化日志记录
- **智能键值对解析**，自动识别 Kitex 风格的日志参数

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "github.com/roc/roc-im-server/tools/klog_otel"
)

func main() {
    ctx := context.Background()
    
    // 初始化 OTEL log provider 和 logger
    logger, lp, err := klog_otel.SetupLogger(ctx, "my-service")
    if err != nil {
        panic(err)
    }
    defer lp.Shutdown(ctx)
    
    // 设置日志级别
    logger.SetLevel(klog_otel.LevelInfo)
    
    // 使用日志
    logger.Info("Service started")
    logger.Infof("Processing request %d", 123)
    
    // 上下文日志
    reqCtx := context.WithValue(ctx, "request_id", "req-456")
    logger.CtxInfof(reqCtx, "Handling request")
}
```

### 在 Kitex 中使用

```go
package main

import (
    "context"
    "github.com/cloudwego/kitex/pkg/klog"
    "github.com/roc/roc-im-server/tools/klog_otel"
)

func main() {
    ctx := context.Background()
    
    // 初始化 OTEL logger
    otelLogger, lp, err := klog_otel.SetupLogger(ctx, "kitex-service")
    if err != nil {
        panic(err)
    }
    defer lp.Shutdown(ctx)
    
    // 设置为 Kitex 的全局 logger
    klog.SetLogger(otelLogger)
    klog.SetLevel(klog.LevelInfo)
    
    // 现在所有 Kitex 的日志都会发送到 OpenTelemetry
    klog.Info("Kitex service started")
    klog.Infof("Listening on port %d", 8080)
}
```

### 使用自定义端点

```go
package main

import (
    "context"
    "github.com/roc/roc-im-server/tools/klog_otel"
)

func main() {
    ctx := context.Background()
    
    // 使用自定义端点初始化
    logger, lp, err := klog_otel.SetupLoggerWithEndpoint(ctx, "my-service", "localhost:4317")
    if err != nil {
        panic(err)
    }
    defer lp.Shutdown(ctx)
    
    logger.Info("Using custom endpoint")
}
```

### 使用配置参数直接构建

```go
package main

import (
    "context"
    "github.com/roc/roc-im-server/tools/klog_otel"
)

func main() {
    ctx := context.Background()
    
    // 使用配置参数直接构建 LoggerProvider 和 logger
    logger, lp, err := klog_otel.SetupLoggerWithConfig(ctx, "my-service", "localhost:4317", true)
    if err != nil {
        panic(err)
    }
    defer lp.Shutdown(ctx)
    
    logger.Info("Using config-based setup")
}
```

### JSON 格式输出

```go
package main

import (
    "context"
    "github.com/roc/roc-im-server/tools/klog_otel"
)

func main() {
    ctx := context.Background()
    
    // 创建 Kitex 兼容的 logger（默认使用 JSON 格式）
    kitexLogger, lp, err := klog_otel.NewKitexLoggerWithConfig(ctx, "my-service", "localhost:4317", true)
    if err != nil {
        panic(err)
    }
    defer lp.Shutdown(ctx)
    
    // 设置为 Kitex 的全局 logger
    klog.SetLogger(kitexLogger)
    klog.SetLevel(klog.LevelInfo)
    
    // 现在所有日志都会以 JSON 格式输出
    klog.Info("Service started")
    
    // 支持键值对参数，会自动解析为 JSON 字段
    klog.CtxInfof(ctx, "sendMessages completed",
        "total_messages", 5,
        "success_count", 3,
        "error_count", 2)
    
    // 输出示例：
    // {"level":"INFO","message":"Service started","timestamp":"1757155653026"}
    // {"level":"INFO","message":"sendMessages completed","timestamp":"1757155653026","fields":{"total_messages":5,"success_count":3,"error_count":2}}
}
```

## API 参考

### 构造函数

#### 基础构造函数
- `NewDefaultLogger(loggerName string) FullLogger` - 使用全局 LoggerProvider 创建 logger
- `NewOtelLogger(loggerName string) *OtelLogger` - 创建基本的 OTEL logger
- `NewOtelLoggerWithProvider(loggerName string, provider *sdklog.LoggerProvider) *OtelLogger` - 使用指定的 LoggerProvider 创建 logger
- `NewOtelLoggerWithConfig(ctx context.Context, loggerName, endpoint string, insecure bool) (*OtelLogger, *sdklog.LoggerProvider, error)` - 使用配置参数构建 LoggerProvider 并创建 logger

#### Kitex 兼容构造函数
- `NewKitexLogger(loggerName string) klog.FullLogger` - 创建与 Kitex 兼容的 logger
- `NewKitexLoggerWithProvider(loggerName string, provider *sdklog.LoggerProvider) klog.FullLogger` - 使用指定的 LoggerProvider 创建 Kitex 兼容的 logger
- `NewKitexLoggerWithConfig(ctx context.Context, loggerName, endpoint string, insecure bool) (klog.FullLogger, *sdklog.LoggerProvider, error)` - 使用配置参数创建 Kitex 兼容的 logger

#### 便捷设置函数
- `SetupLogger(ctx context.Context, loggerName string) (FullLogger, *sdklog.LoggerProvider, error)` - 完整的设置流程
- `SetupLoggerWithEndpoint(ctx context.Context, loggerName, endpoint string) (FullLogger, *sdklog.LoggerProvider, error)` - 使用指定端点设置
- `SetupLoggerWithConfig(ctx context.Context, loggerName, endpoint string, insecure bool) (FullLogger, *sdklog.LoggerProvider, error)` - 使用配置参数设置
- `SetupLoggerWithInsecureEndpoint(ctx context.Context, loggerName, endpoint string) (FullLogger, *sdklog.LoggerProvider, error)` - 使用不安全的端点设置
- `NewKitexCompatibleLogger(loggerName string) klog.FullLogger` - 创建与 Kitex 兼容的 logger

### 日志级别

```go
const (
    LevelTrace Level = iota
    LevelDebug
    LevelInfo
    LevelNotice
    LevelWarn
    LevelError
    LevelFatal
)
```

### 主要方法

#### 基本日志方法
- `Trace(v ...interface{})`
- `Debug(v ...interface{})`
- `Info(v ...interface{})`
- `Notice(v ...interface{})`
- `Warn(v ...interface{})`
- `Error(v ...interface{})`
- `Fatal(v ...interface{})`

#### 格式化日志方法
- `Tracef(format string, v ...interface{})`
- `Debugf(format string, v ...interface{})`
- `Infof(format string, v ...interface{})`
- `Noticef(format string, v ...interface{})`
- `Warnf(format string, v ...interface{})`
- `Errorf(format string, v ...interface{})`
- `Fatalf(format string, v ...interface{})`

#### 上下文日志方法
- `CtxTracef(ctx context.Context, format string, v ...interface{})`
- `CtxDebugf(ctx context.Context, format string, v ...interface{})`
- `CtxInfof(ctx context.Context, format string, v ...interface{})`
- `CtxNoticef(ctx context.Context, format string, v ...interface{})`
- `CtxWarnf(ctx context.Context, format string, v ...interface{})`
- `CtxErrorf(ctx context.Context, format string, v ...interface{})`
- `CtxFatalf(ctx context.Context, format string, v ...interface{})`

#### 控制方法
- `SetLevel(level Level)` - 设置日志级别
- `SetOutput(w io.Writer)` - 设置输出（在 OTEL 实现中此方法被忽略）
- `SetJSONFormat(useJSON bool)` - 设置是否使用 JSON 格式输出
- `IsJSONFormat() bool` - 返回是否使用 JSON 格式

## JSON 格式输出

### 自动键值对解析

klog_otel 会自动识别 Kitex 风格的日志参数，将键值对转换为 JSON 字段：

```go
// 输入
klog.CtxInfof(ctx, "sendMessages completed",
    "total_messages", 5,
    "success_count", 3,
    "error_count", 2)

// 输出
{
    "level": "INFO",
    "message": "sendMessages completed",
    "timestamp": "1757155653026",
    "fields": {
        "total_messages": 5,
        "success_count": 3,
        "error_count": 2
    }
}
```

### 格式控制

```go
// 启用 JSON 格式（默认）
logger.SetJSONFormat(true)

// 禁用 JSON 格式，使用原始格式
logger.SetJSONFormat(false)

// 检查当前格式
isJSON := logger.IsJSONFormat()
```

## 级别映射

| klog Level | slog Level | 说明 |
|------------|------------|------|
| Trace      | Debug      | 跟踪信息 |
| Debug      | Debug      | 调试信息 |
| Info       | Info       | 一般信息 |
| Notice     | Info       | 重要通知 |
| Warn       | Warn       | 警告信息 |
| Error      | Error      | 错误信息 |
| Fatal      | Error      | 致命错误 |

## 注意事项

1. 所有日志都会发送到 OpenTelemetry 后端，不会输出到本地文件或控制台
2. 需要确保 OpenTelemetry 收集器正在运行并可以接收日志
3. 在生产环境中，建议设置合适的日志级别以减少日志量
4. 使用完毕后请调用 `lp.Shutdown(ctx)` 来优雅关闭 LoggerProvider

## 依赖

- `go.opentelemetry.io/contrib/bridges/otelslog`
- `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc`
- `go.opentelemetry.io/otel/log/global`
- `go.opentelemetry.io/otel/sdk/log`
