# Go encoding/json 使用范例

本目录包含了 Go 语言 `encoding/json` 包的详细使用范例。

## 运行示例

```bash
# 进入示例目录
cd examples

# 运行完整示例
go run json_example.go
```

## 主要功能演示

### 1. 基本操作
- **结构体转JSON** - `json.Marshal()`
- **JSON转结构体** - `json.Unmarshal()`
- **格式化输出** - `json.MarshalIndent()`

### 2. JSON 标签
- `json:"field_name"` - 自定义字段名
- `json:"field_name,omitempty"` - 空值时忽略
- `json:"-"` - 忽略字段

### 3. 数据类型处理
- **数组和切片** - `[]Type`
- **映射表** - `map[string]interface{}`
- **嵌套结构体** - 复杂数据结构
- **指针类型** - `*Type`

### 4. 高级特性
- **自定义编解码** - 实现 `MarshalJSON` 和 `UnmarshalJSON` 接口
- **流式处理** - 使用 `json.Encoder` 和 `json.Decoder`
- **文件读写** - 配置文件处理

### 5. 实际应用场景
- **API 响应格式** - 标准的 REST API 响应
- **配置文件** - 应用程序配置管理
- **数据传输** - 网络数据交换

## 常用技巧

### 时间处理
```go
// 使用 time.Time 类型
type User struct {
    CreateAt time.Time `json:"create_at"`
}

// 自定义时间格式
type CustomTime struct {
    Time time.Time
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(ct.Time.Format("2006-01-02 15:04:05"))
}
```

### 可选字段
```go
type User struct {
    ID    int     `json:"id"`
    Email string  `json:"email,omitempty"`    // 空值时忽略
    Age   *int    `json:"age,omitempty"`      // 指针类型，nil时忽略
}
```

### 忽略字段
```go
type User struct {
    ID       int    `json:"id"`
    Password string `json:"-"`                // 完全忽略
}
```

## 注意事项

1. **数字类型** - JSON 中的数字默认解析为 `float64`
2. **空值处理** - 使用 `omitempty` 标签处理空值
3. **字段命名** - 使用 JSON 标签指定字段名
4. **类型断言** - 使用 `interface{}` 时需要类型断言
5. **错误处理** - 始终检查编解码错误

## 性能建议

- 对于大量数据，使用流式处理 (`json.Encoder`/`json.Decoder`)
- 避免使用 `interface{}` 类型，明确定义结构体
- 使用结构体标签优化 JSON 输出
- 考虑使用 `json.RawMessage` 延迟解析 