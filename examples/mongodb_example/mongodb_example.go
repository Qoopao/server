package mongodb_example

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/roc/roc-im-server/tools/db"
	"github.com/roc/roc-im-server/tools/log"
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
	FromUserID string `bson:"from_user_id" json:"from_user_id"`
	ToUserID   string `bson:"to_user_id" json:"to_user_id"`
	Content    string `bson:"content" json:"content"`
	Type       string `bson:"type" json:"type"`     // text, image, file
	Status     string `bson:"status" json:"status"` // sent, delivered, read
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

	// 测试核心接口覆盖
	testAdvancedOps(ctx, mongoDB, userCollection, messageCollection, logger)

	// 测试数据库级索引与负例
	testDBLevelIndexesAndNegatives(ctx, mongoDB, userCollection, logger)

	logger.Info("=== MongoDB 示例完成 ===")
}

// testAdvancedOps 覆盖更多核心接口与错误日志
func testAdvancedOps(ctx context.Context, mongoDB db.Database, users db.Collection, messages db.Collection, logger log.Logger) {
	logger.Info("--- 测试核心接口（高级操作） ---")

	// 1) FindOne
	one := users.FindOne(ctx, map[string]any{"is_active": true})
	var oneUser User
	if err := one.Decode(&oneUser); err != nil {
		logger.Error("FindOne 解码失败", log.Error(err))
	} else {
		logger.Info("✓ FindOne 成功", log.Any("user", oneUser.Email))
	}

	// 2) FindOneAndUpdate（after）
	fou := users.FindOneAndUpdate(ctx,
		map[string]any{"email": oneUser.Email},
		map[string]any{"$set": map[string]any{"age": oneUser.Age + 1}},
		&db.FindOneAndUpdateOptions{ReturnDocument: "after", Upsert: false},
	)
	var afterUpdate User
	if err := fou.Decode(&afterUpdate); err != nil {
		logger.Error("FindOneAndUpdate 失败", log.Error(err))
	} else {
		logger.Info("✓ FindOneAndUpdate 成功", log.Int("age", afterUpdate.Age))
	}

	// 3) FindOneAndReplace
	replaceUser := afterUpdate
	replaceUser.IsActive = afterUpdate.IsActive
	replaceUser.Name = "替换后的用户"
	forReplace := users.FindOneAndReplace(ctx,
		map[string]any{"email": afterUpdate.Email},
		&replaceUser,
		&db.FindOneAndReplaceOptions{Upsert: false},
	)
	var replaced User
	if err := forReplace.Decode(&replaced); err != nil {
		logger.Error("FindOneAndReplace 失败", log.Error(err))
	} else {
		logger.Info("✓ FindOneAndReplace 成功")
	}

	// 4) FindOneAndDelete（演示，不影响主数据：用一个几乎不命中的条件）
	fod := users.FindOneAndDelete(ctx, map[string]any{"email": "not-exist@example.com"})
	var deleted User
	if err := fod.Decode(&deleted); err != nil {
		logger.Info("FindOneAndDelete 未找到预期记录（正常）")
	} else {
		logger.Info("✓ FindOneAndDelete 成功（意外找到记录）")
	}

	// 5) ReplaceOne
	replaceUser2 := replaced
	replaceUser2.Name = "ReplaceOne 用户"
	if _, err := users.ReplaceOne(ctx, map[string]any{"email": replaceUser2.Email}, &replaceUser2); err != nil {
		logger.Error("ReplaceOne 失败", log.Error(err))
	} else {
		logger.Info("✓ ReplaceOne 成功")
	}

	// 6) DeleteOne / DeleteMany（仅演示，不删除真实数据）
	if _, err := users.DeleteOne(ctx, map[string]any{"email": "not-exist@example.com"}); err != nil {
		logger.Error("DeleteOne 失败", log.Error(err))
	} else {
		logger.Info("✓ DeleteOne 执行完成（可能未删除）")
	}
	if _, err := users.DeleteMany(ctx, map[string]any{"email": map[string]any{"$regex": "^not-exist-"}}); err != nil {
		logger.Error("DeleteMany 失败", log.Error(err))
	} else {
		logger.Info("✓ DeleteMany 执行完成（可能未删除）")
	}

	// 7) EstimatedDocumentCount
	if n, err := messages.EstimatedDocumentCount(ctx); err != nil {
		logger.Error("EstimatedDocumentCount 失败", log.Error(err))
	} else {
		logger.Info("✓ EstimatedDocumentCount 成功", log.Int64("count", n))
	}

	// 8) Distinct（用户邮箱）
	if values, err := users.Distinct(ctx, "email", map[string]any{}); err != nil {
		logger.Error("Distinct 失败", log.Error(err))
	} else {
		logger.Info("✓ Distinct 成功", log.Int("emails", len(values)))
	}

	// 9) 集合级索引：CreateIndexes / DropIndex / ListIndexes
	collIdxModels := []db.IndexModel{
		{Keys: []db.KeyValue{{Key: "name", Value: 1}}},
		{Keys: []db.KeyValue{{Key: "age", Value: -1}}},
	}
	names, err := users.CreateIndexes(ctx, collIdxModels)
	if err != nil {
		logger.Error("Collection.CreateIndexes 失败", log.Error(err))
	} else {
		logger.Info("✓ Collection.CreateIndexes 成功", log.Any("names", names))
		if len(names) > 0 {
			if err := users.DropIndex(ctx, names[0]); err != nil {
				logger.Error("Collection.DropIndex 失败", log.Error(err))
			} else {
				logger.Info("✓ Collection.DropIndex 成功", log.Any("name", names[0]))
			}
		}
	}
	// ListIndexes（集合级）
	if cur, err := users.ListIndexes(ctx); err != nil {
		logger.Error("Collection.ListIndexes 失败", log.Error(err))
	} else {
		var idxs []map[string]any
		if err := cur.All(ctx, &idxs); err != nil {
			logger.Error("Collection.ListIndexes 解析失败", log.Error(err))
		} else {
			logger.Info("✓ Collection.ListIndexes 成功", log.Int("count", len(idxs)))
		}
	}

	// 10) 集合 Drop：创建临时集合，插入后删除
	temp := mongoDB.GetCollection("tmp_demo")
	if _, err := temp.InsertOne(ctx, map[string]any{"k": "v"}); err != nil {
		logger.Error("临时集合插入失败", log.Error(err))
	} else {
		logger.Info("✓ 临时集合插入成功")
	}
	if err := temp.Drop(ctx); err != nil {
		logger.Error("临时集合 Drop 失败", log.Error(err))
	} else {
		logger.Info("✓ 临时集合 Drop 成功")
	}
}

// testDBLevelIndexesAndNegatives 数据库级索引与负例覆盖
func testDBLevelIndexesAndNegatives(ctx context.Context, mongoDB db.Database, users db.Collection, logger log.Logger) {
	logger.Info("--- 测试数据库级索引与负例 ---")

	// 数据库级：CreateIndexes / DropIndex / ListIndexes
	models := []db.IndexModel{
		{Keys: []db.KeyValue{{Key: "is_active", Value: 1}}},
		{Keys: []db.KeyValue{{Key: "age", Value: 1}}},
	}
	if err := mongoDB.CreateIndexes(ctx, "users", models); err != nil {
		logger.Error("DB.CreateIndexes 失败", log.Error(err))
	} else {
		logger.Info("✓ DB.CreateIndexes 成功")
	}
	// DropIndex（尝试删除一个可能存在的索引）
	if err := mongoDB.DropIndex(ctx, "users", "is_active_1"); err != nil {
		logger.Error("DB.DropIndex 失败（可能索引不存在）", log.Error(err))
	} else {
		logger.Info("✓ DB.DropIndex 成功", log.String("name", "is_active_1"))
	}
	if idxs, err := mongoDB.ListIndexes(ctx, "users"); err != nil {
		logger.Error("DB.ListIndexes 失败", log.Error(err))
	} else {
		logger.Info("✓ DB.ListIndexes 成功", log.Int("count", len(idxs)))
	}

	// 负例：FindOne 查不到（应返回解码错误）
	notFound := users.FindOne(ctx, map[string]any{"email": "not-exist-xyz@example.com"})
	var nf User
	if err := notFound.Decode(&nf); err != nil {
		logger.Info("✓ FindOne 未找到返回错误（预期）")
	} else {
		logger.Error("FindOne 负例失败：意外找到记录")
	}

	// 负例：Aggregate 非法管道
	badPipeline := []map[string]any{{"$unknown": map[string]any{"x": 1}}}
	if cur, err := users.Aggregate(ctx, badPipeline); err != nil {
		logger.Info("✓ Aggregate 非法管道返回错误（预期）")
	} else {
		_ = cur.Close(ctx)
		logger.Error("Aggregate 负例失败：非法管道未报错")
	}
}

// testUserOperations 测试用户操作
func testUserOperations(ctx context.Context, collection db.Collection, logger log.Logger) {
	logger.Info("--- 测试用户操作 ---")

	// 创建用户
	user := &User{
		BaseEntity: db.BaseEntity{},
		Name:       "张三",
		Email:      fmt.Sprintf("zhangsan+%d@example.com", time.Now().UnixNano()),
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
			Email:      fmt.Sprintf("lisi+%d@example.com", time.Now().UnixNano()),
			Age:        30,
			IsActive:   true,
		},
		&User{
			BaseEntity: db.BaseEntity{},
			Name:       "王五",
			Email:      fmt.Sprintf("wangwu+%d@example.com", time.Now().UnixNano()+1),
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
	filter := map[string]any{"is_active": true}
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
	update := map[string]any{
		"$set": map[string]any{
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
	count, err := collection.CountDocuments(ctx, map[string]any{})
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
		FromUserID: fmt.Sprintf("%d", time.Now().UnixNano()),
		ToUserID:   fmt.Sprintf("%d", time.Now().UnixNano()+1),
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
	filter := map[string]any{"type": "text"}
	opts := &db.FindOptions{Sort: map[string]any{"created_at": -1}, Limit: func(v int64) *int64 { return &v }(10)}
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
	updateFilter := map[string]any{"status": "sent"}
	update := map[string]any{
		"$set": map[string]any{
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
	pipeline := []map[string]any{
		{
			"$group": map[string]any{
				"_id":   "$type",
				"count": map[string]any{"$sum": 1},
			},
		},
		{
			"$sort": []db.KeyValue{
				{Key: "count", Value: -1},
				{Key: "_id", Value: 1},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("聚合查询失败", log.Error(err))
		return
	}
	defer cursor.Close(ctx)

	var results []map[string]any
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
		Email:      fmt.Sprintf("transaction+%d@example.com", time.Now().UnixNano()),
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
		FromUserID: fmt.Sprint(userResult.InsertedID),
		ToUserID:   fmt.Sprintf("%d", time.Now().UnixNano()+2),
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
	emailUnique := true
	indexModel := db.IndexModel{
		Keys:    []db.KeyValue{{Key: "email", Value: 1}},
		Options: &db.IndexOptions{Unique: &emailUnique},
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
	bg := true
	messageIndexModel := db.IndexModel{
		Keys: []db.KeyValue{
			{Key: "from_user_id", Value: 1},
			{Key: "to_user_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
		Options: &db.IndexOptions{Background: &bg},
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
