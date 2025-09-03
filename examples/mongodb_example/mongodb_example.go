package mongodb_example

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/roc/roc-im-server/tools/db"
	"github.com/roc/roc-im-server/tools/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 用户实体示例
type User struct {
	db.BaseEntity
	Name     string `bson:"name" json:"name"`
	Email    string `bson:"email" json:"email"`
	Age      int    `bson:"age" json:"age"`
	IsActive bool   `bson:"is_active" json:"is_active"`
}

// Message 消息实体示例
type Message struct {
	db.BaseEntity
	FromUserID primitive.ObjectID `bson:"from_user_id" json:"from_user_id"`
	ToUserID   primitive.ObjectID `bson:"to_user_id" json:"to_user_id"`
	Content    string             `bson:"content" json:"content"`
	Type       string             `bson:"type" json:"type"`     // text, image, file
	Status     string             `bson:"status" json:"status"` // sent, delivered, read
}

// MongoDBExample_main MongoDB示例主函数
func MongoDBExample_main() {
	// 创建日志配置
	logConfig := log.ProductionConfig()
	logConfig.OutputPaths = []string{
		"logs/mongodb_example.log", // 输出到文件
	}

	// 创建日志器
	logger, err := log.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("创建日志器失败: %v", err))
	}
	defer logger.Sync()

	logger.Info("=== MongoDB 示例开始 ===")

	// 创建连接配置
	config := db.DefaultConnectionConfig()
	config.Database = "roc_im_example"
	// 根据 docker/mongodb/docker-compose.yml 配置的认证信息
	config.Username = "admin"      // 根用户
	config.Password = "mongodb123" // 根密码
	config.AuthSource = "admin"    // 认证数据库

	// 创建MongoDB实例
	mongoDB := db.NewMongoDB(config)

	// 连接数据库
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoDB.Connect(ctx); err != nil {
		logger.Fatal("连接MongoDB失败", log.Error(err))
	}
	defer mongoDB.Disconnect(ctx)

	// 测试连接
	if err := mongoDB.Ping(ctx); err != nil {
		logger.Fatal("Ping MongoDB失败", log.Error(err))
	}
	logger.Info("✓ MongoDB连接成功")

	// 获取集合
	userCollection := mongoDB.GetCollection("users")
	messageCollection := mongoDB.GetCollection("messages")

	// 测试用户操作
	testUserOperations(ctx, userCollection, logger)

	// 测试消息操作
	testMessageOperations(ctx, messageCollection, logger)

	// 测试聚合查询
	testAggregation(ctx, messageCollection, logger)

	// 测试事务
	testTransaction(ctx, mongoDB, logger)

	// 测试索引
	testIndexes(ctx, mongoDB, logger)

	logger.Info("=== MongoDB 示例完成 ===")
}

// testUserOperations 测试用户操作
func testUserOperations(ctx context.Context, collection db.Collection, logger log.Logger) {
	logger.Info("--- 测试用户操作 ---")

	// 创建用户
	user := &User{
		BaseEntity: db.BaseEntity{},
		Name:       "张三",
		Email:      "zhangsan@example.com",
		Age:        25,
		IsActive:   true,
	}
	user.BeforeCreate()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		logger.Error("插入用户失败", log.Error(err))
		return
	}
	logger.Info("✓ 创建用户成功", log.Any("ID", result.InsertedID))

	// 批量创建用户
	users := []interface{}{
		&User{
			BaseEntity: db.BaseEntity{},
			Name:       "李四",
			Email:      "lisi@example.com",
			Age:        30,
			IsActive:   true,
		},
		&User{
			BaseEntity: db.BaseEntity{},
			Name:       "王五",
			Email:      "wangwu@example.com",
			Age:        28,
			IsActive:   false,
		},
	}

	// 设置创建时间
	for _, u := range users {
		if user, ok := u.(*User); ok {
			user.BeforeCreate()
		}
	}

	_, err = collection.InsertMany(ctx, users)
	if err != nil {
		logger.Error("批量插入用户失败", log.Error(err))
		return
	}
	logger.Info("✓ 批量创建用户成功")

	// 查询用户
	filter := bson.M{"is_active": true}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		logger.Error("查询用户失败", log.Error(err))
		return
	}
	defer cursor.Close(ctx)

	var activeUsers []User
	if err = cursor.All(ctx, &activeUsers); err != nil {
		logger.Error("解析用户数据失败", log.Error(err))
		return
	}
	logger.Info("✓ 查询到活跃用户", log.Int("count", len(activeUsers)))

	// 更新用户
	update := bson.M{
		"$set": bson.M{
			"age":        26,
			"updated_at": time.Now(),
		},
	}
	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("更新用户失败", log.Error(err))
		return
	}
	logger.Info("✓ 更新了用户", log.Int64("count", updateResult.ModifiedCount))

	// 统计用户数量
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Error("统计用户数量失败", log.Error(err))
		return
	}
	logger.Info("✓ 总用户数", log.Int64("count", count))
}

// testMessageOperations 测试消息操作
func testMessageOperations(ctx context.Context, collection db.Collection, logger log.Logger) {
	logger.Info("--- 测试消息操作 ---")

	// 创建消息
	message := &Message{
		BaseEntity: db.BaseEntity{},
		FromUserID: primitive.NewObjectID(),
		ToUserID:   primitive.NewObjectID(),
		Content:    "Hello, World!",
		Type:       "text",
		Status:     "sent",
	}
	message.BeforeCreate()

	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		logger.Error("插入消息失败", log.Error(err))
		return
	}
	fmt.Printf("✓ 创建消息成功，ID: %v\n", result.InsertedID)

	// 查询消息
	filter := bson.M{"type": "text"}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(10)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("查询消息失败", log.Error(err))
		return
	}
	defer cursor.Close(ctx)

	var messages []Message
	if err = cursor.All(ctx, &messages); err != nil {
		logger.Error("解析消息数据失败", log.Error(err))
		return
	}
	fmt.Printf("✓ 查询到 %d 条文本消息\n", len(messages))

	// 更新消息状态
	updateFilter := bson.M{"status": "sent"}
	update := bson.M{
		"$set": bson.M{
			"status":     "delivered",
			"updated_at": time.Now(),
		},
	}
	updateResult, err := collection.UpdateMany(ctx, updateFilter, update)
	if err != nil {
		logger.Error("更新消息状态失败", log.Error(err))
		return
	}
	fmt.Printf("✓ 更新了 %d 条消息状态\n", updateResult.ModifiedCount)
}

// testAggregation 测试聚合查询
func testAggregation(ctx context.Context, collection db.Collection, logger log.Logger) {
	logger.Info("--- 测试聚合查询 ---")

	// 按类型统计消息数量
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("聚合查询失败", log.Error(err))
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		logger.Error("解析聚合结果失败", log.Error(err))
		return
	}

	fmt.Println("✓ 消息类型统计:")
	for _, result := range results {
		fmt.Printf("  类型: %v, 数量: %v\n", result["_id"], result["count"])
	}
}

// testTransaction 测试事务
func testTransaction(ctx context.Context, mongoDB db.Database, logger log.Logger) {
	logger.Info("--- 测试事务 ---")
	logger.Info("注意：单机MongoDB不支持事务，需要副本集或分片集群")

	// 模拟事务操作（实际环境中需要副本集）
	userCollection := mongoDB.GetCollection("users")
	messageCollection := mongoDB.GetCollection("messages")

	// 创建用户
	user := &User{
		BaseEntity: db.BaseEntity{},
		Name:       "事务用户",
		Email:      "transaction@example.com",
		Age:        35,
		IsActive:   true,
	}
	user.BeforeCreate()

	userResult, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		logger.Error("创建用户失败", log.Error(err))
		return
	}

	// 创建消息
	message := &Message{
		BaseEntity: db.BaseEntity{},
		FromUserID: userResult.InsertedID.(primitive.ObjectID),
		ToUserID:   primitive.NewObjectID(),
		Content:    "模拟事务消息",
		Type:       "text",
		Status:     "sent",
	}
	message.BeforeCreate()

	_, err = messageCollection.InsertOne(ctx, message)
	if err != nil {
		logger.Error("创建消息失败", log.Error(err))
		return
	}

	logger.Info("✓ 模拟事务操作成功（实际需要副本集支持真正的事务）")
}

// testIndexes 测试索引
func testIndexes(ctx context.Context, mongoDB db.Database, logger log.Logger) {
	logger.Info("--- 测试索引 ---")

	// 创建用户邮箱唯一索引（检查是否已存在）
	emailIndex := options.Index().SetUnique(true)
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: emailIndex,
	}

	err := mongoDB.CreateIndex(ctx, "users", indexModel)
	if err != nil {
		// 如果是重复索引错误，说明索引已存在，这是正常的
		if strings.Contains(err.Error(), "duplicate key error") || strings.Contains(err.Error(), "already exists") {
			logger.Info("✓ 邮箱唯一索引已存在")
		} else {
			logger.Error("创建邮箱索引失败", log.Error(err))
		}
	} else {
		logger.Info("✓ 创建邮箱唯一索引成功")
	}

	// 创建消息复合索引
	messageIndex := options.Index().SetBackground(true)
	messageIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "from_user_id", Value: 1},
			{Key: "to_user_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
		Options: messageIndex,
	}

	err = mongoDB.CreateIndex(ctx, "messages", messageIndexModel)
	if err != nil {
		// 如果是重复索引错误，说明索引已存在，这是正常的
		if strings.Contains(err.Error(), "duplicate key error") || strings.Contains(err.Error(), "already exists") {
			logger.Info("✓ 消息复合索引已存在")
		} else {
			logger.Error("创建消息复合索引失败", log.Error(err))
		}
	} else {
		logger.Info("✓ 创建消息复合索引成功")
	}

	// 列出索引
	indexes, err := mongoDB.ListIndexes(ctx, "users")
	if err != nil {
		logger.Error("列出索引失败", log.Error(err))
		return
	}

	logger.Info("✓ 用户集合索引数量", log.Int("count", len(indexes)))
}
