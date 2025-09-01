package examples

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// User 用户结构体
type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Age      int       `json:"age"`
	IsActive bool      `json:"is_active"`
	CreateAt time.Time `json:"create_at"`
}

// UserWithTags 带标签的用户结构体
type UserWithTags struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	Email    string     `json:"email,omitempty"` // 空值时忽略
	Age      int        `json:"age"`
	Password string     `json:"-"` // 忽略此字段
	IsActive bool       `json:"is_active"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt *time.Time `json:"update_at,omitempty"` // 指针类型，空值时忽略
}

// Group 群组结构体
type Group struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Members     []User                 `json:"members"`
	Settings    map[string]interface{} `json:"settings"`
}

// Message 消息结构体
type Message struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	Sender     User                   `json:"sender"`
	Recipients []User                 `json:"recipients"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// JSON_main JSON示例主函数
func JSON_main() {
	fmt.Println("=== Go encoding/json 使用范例 ===\n")

	// 1. 结构体转JSON (Marshal)
	example1_StructToJSON()

	// 2. JSON转结构体 (Unmarshal)
	example2_JSONToStruct()

	// 3. 处理JSON标签
	example3_JSONTags()

	// 4. 处理map和interface{}
	example4_MapAndInterface()

	// 5. 处理JSON数组
	example5_JSONArray()

	// 6. 嵌套结构体
	example6_NestedStruct()

	// 7. 自定义JSON编解码
	example7_CustomJSON()

	// 8. 流式JSON处理
	example8_StreamJSON()

	// 9. 实际应用场景
	example9_RealWorldUsage()
}

// 1. 结构体转JSON
func example1_StructToJSON() {
	fmt.Println("1. 结构体转JSON (Marshal)")
	fmt.Println("---")

	user := User{
		ID:       1,
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Age:      30,
		IsActive: true,
		CreateAt: time.Now(),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("JSON: %s\n", jsonData)

	// 格式化输出
	prettyJSON, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("格式化JSON:\n%s\n\n", prettyJSON)
}

// 2. JSON转结构体
func example2_JSONToStruct() {
	fmt.Println("2. JSON转结构体 (Unmarshal)")
	fmt.Println("---")

	jsonStr := `{
		"id": 2,
		"name": "李四",
		"email": "lisi@example.com",
		"age": 25,
		"is_active": true,
		"create_at": "2023-01-01T00:00:00Z"
	}`

	var user User
	err := json.Unmarshal([]byte(jsonStr), &user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("解析后的结构体: %+v\n\n", user)
}

// 3. 处理JSON标签
func example3_JSONTags() {
	fmt.Println("3. 处理JSON标签")
	fmt.Println("---")

	user := UserWithTags{
		ID:       3,
		Name:     "王五",
		Email:    "", // 空值，会被omitempty忽略
		Age:      28,
		Password: "secret123", // 会被"-"标签忽略
		IsActive: true,
		CreateAt: time.Now(),
		UpdateAt: nil, // nil指针，会被omitempty忽略
	}

	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("带标签的JSON:\n%s\n\n", jsonData)
}

// 4. 处理map和interface{}
func example4_MapAndInterface() {
	fmt.Println("4. 处理map和interface{}")
	fmt.Println("---")

	// JSON to map
	jsonStr := `{
		"name": "test",
		"age": 30,
		"scores": [85, 90, 95],
		"address": {
			"city": "北京",
			"country": "中国"
		}
	}`

	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("解析到map: %+v\n", data)
	fmt.Printf("姓名: %s\n", data["name"])
	fmt.Printf("年龄: %.0f\n", data["age"].(float64)) // JSON数字默认为float64
	fmt.Printf("分数: %v\n", data["scores"])

	// 访问嵌套对象
	address := data["address"].(map[string]interface{})
	fmt.Printf("城市: %s\n\n", address["city"])
}

// 5. 处理JSON数组
func example5_JSONArray() {
	fmt.Println("5. 处理JSON数组")
	fmt.Println("---")

	users := []User{
		{ID: 1, Name: "用户1", Email: "user1@example.com", Age: 25, IsActive: true, CreateAt: time.Now()},
		{ID: 2, Name: "用户2", Email: "user2@example.com", Age: 30, IsActive: false, CreateAt: time.Now()},
		{ID: 3, Name: "用户3", Email: "user3@example.com", Age: 35, IsActive: true, CreateAt: time.Now()},
	}

	// 数组转JSON
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("用户数组JSON:\n%s\n", jsonData)

	// JSON转数组
	var parsedUsers []User
	err = json.Unmarshal(jsonData, &parsedUsers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("解析后的用户数组: %d 个用户\n\n", len(parsedUsers))
}

// 6. 嵌套结构体
func example6_NestedStruct() {
	fmt.Println("6. 嵌套结构体")
	fmt.Println("---")

	group := Group{
		ID:          "group_001",
		Name:        "开发团队",
		Description: "软件开发团队",
		Members: []User{
			{ID: 1, Name: "开发者1", Email: "dev1@example.com", Age: 28, IsActive: true, CreateAt: time.Now()},
			{ID: 2, Name: "开发者2", Email: "dev2@example.com", Age: 32, IsActive: true, CreateAt: time.Now()},
		},
		Settings: map[string]interface{}{
			"max_members": 10,
			"public":      true,
			"tags":        []string{"开发", "技术"},
		},
	}

	jsonData, err := json.MarshalIndent(group, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("群组JSON:\n%s\n\n", jsonData)
}

// CustomTime 自定义时间格式
type CustomTime struct {
	Time time.Time
}

// MarshalJSON 实现MarshalJSON接口
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Time.Format("2006-01-02 15:04:05"))
}

// UnmarshalJSON 实现UnmarshalJSON接口
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// Event 事件结构体
type Event struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	StartTime CustomTime `json:"start_time"`
}

// APIResponse API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Config 配置文件格式
type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	} `json:"redis"`
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
}

// 7. 自定义JSON编解码
func example7_CustomJSON() {
	fmt.Println("7. 自定义JSON编解码")
	fmt.Println("---")

	event := Event{
		ID:        "event_001",
		Name:      "会议",
		StartTime: CustomTime{Time: time.Now()},
	}

	jsonData, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("自定义时间格式JSON:\n%s\n\n", jsonData)
}

// 8. 流式JSON处理
func example8_StreamJSON() {
	fmt.Println("8. 流式JSON处理")
	fmt.Println("---")

	// 创建JSON数据
	users := []User{
		{ID: 1, Name: "流式用户1", Email: "stream1@example.com", Age: 25, IsActive: true, CreateAt: time.Now()},
		{ID: 2, Name: "流式用户2", Email: "stream2@example.com", Age: 30, IsActive: false, CreateAt: time.Now()},
	}

	// 使用Encoder写入
	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")

	for _, user := range users {
		if err := encoder.Encode(user); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("流式编码结果:\n%s\n", buf.String())

	// 使用Decoder读取
	decoder := json.NewDecoder(strings.NewReader(buf.String()))

	var decodedUsers []User
	for decoder.More() {
		var user User
		if err := decoder.Decode(&user); err != nil {
			log.Fatal(err)
		}
		decodedUsers = append(decodedUsers, user)
	}

	fmt.Printf("流式解码结果: %d 个用户\n\n", len(decodedUsers))
}

// 9. 实际应用场景
func example9_RealWorldUsage() {
	fmt.Println("9. 实际应用场景")
	fmt.Println("---")

	// 创建API响应
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "用户1"},
				{"id": 2, "name": "用户2"},
			},
			"total": 2,
		},
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("API响应JSON:\n%s\n", responseJSON)

	// 创建配置文件
	config := Config{}
	config.Database.Host = "localhost"
	config.Database.Port = 3306
	config.Database.Username = "root"
	config.Database.Password = "password"
	config.Redis.Host = "localhost"
	config.Redis.Port = 6379
	config.Redis.Password = "redis_password"
	config.Server.Host = "0.0.0.0"
	config.Server.Port = 8080

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("配置文件JSON:\n%s\n", configJSON)

	// 保存配置到文件
	if err := saveConfigToFile("config.json", config); err != nil {
		log.Printf("保存配置文件失败: %v\n", err)
	} else {
		fmt.Println("配置文件已保存到 config.json")
	}

	// 从文件读取配置
	if loadedConfig, err := loadConfigFromFile("config.json"); err != nil {
		log.Printf("读取配置文件失败: %v\n", err)
	} else {
		fmt.Printf("从文件读取的配置: %+v\n", loadedConfig.Database)
	}
}

// 保存配置到文件
func saveConfigToFile(filename string, config Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

// 从文件加载配置
func loadConfigFromFile(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

// 常用的JSON处理函数

// PrettyPrint 格式化打印JSON
func PrettyPrint(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("JSON格式化失败: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}

// JSONToString 将对象转换为JSON字符串
func JSONToString(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

// StringToJSON 将JSON字符串转换为对象
func StringToJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// IsValidJSON 检查字符串是否为有效JSON
func IsValidJSON(jsonStr string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}
