package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB 实现 Database 接口
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   *ConnectionConfig
}

// NewMongoDB 创建新的 MongoDB 实例
func NewMongoDB(config *ConnectionConfig) *MongoDB {
	return &MongoDB{
		config: config,
	}
}

// Connect 连接到 MongoDB
func (m *MongoDB) Connect(ctx context.Context) error {
	clientOptions := options.Client().
		ApplyURI(m.config.URI).
		SetMaxPoolSize(m.config.MaxPoolSize).
		SetMinPoolSize(m.config.MinPoolSize).
		SetMaxConnIdleTime(m.config.MaxIdleTime).
		SetConnectTimeout(m.config.ConnectTimeout).
		SetSocketTimeout(m.config.SocketTimeout).
		SetServerSelectionTimeout(m.config.ServerTimeout)

	if m.config.Username != "" && m.config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			AuthSource: m.config.AuthSource,
			Username:   m.config.Username,
			Password:   m.config.Password,
		})
	}

	if m.config.SSL {
		clientOptions.SetTLSConfig(nil) // 使用默认 TLS 配置
	}

	if m.config.ReplicaSet != "" {
		clientOptions.SetReplicaSet(m.config.ReplicaSet)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)

	return nil
}

// Disconnect 断开连接
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

// Ping 检查连接状态
func (m *MongoDB) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("client is not connected")
	}
	return m.client.Ping(ctx, nil)
}

// GetDatabase 获取数据库
func (m *MongoDB) GetDatabase(name string) Database {
	return &MongoDB{
		client:   m.client,
		database: m.client.Database(name),
		config:   m.config,
	}
}

// GetCollection 获取集合
func (m *MongoDB) GetCollection(name string) Collection {
	return &MongoCollection{
		collection: m.database.Collection(name),
	}
}

// WithTransaction 执行事务
func (m *MongoDB) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return nil, fn(sc)
	})

	return err
}

// CreateIndex 创建索引
func (m *MongoDB) CreateIndex(ctx context.Context, collection string, model mongo.IndexModel) error {
	_, err := m.database.Collection(collection).Indexes().CreateOne(ctx, model)
	return err
}

// CreateIndexes 创建多个索引
func (m *MongoDB) CreateIndexes(ctx context.Context, collection string, models []mongo.IndexModel) error {
	_, err := m.database.Collection(collection).Indexes().CreateMany(ctx, models)
	return err
}

// DropIndex 删除索引
func (m *MongoDB) DropIndex(ctx context.Context, collection string, name string) error {
	_, err := m.database.Collection(collection).Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes 列出索引
func (m *MongoDB) ListIndexes(ctx context.Context, collection string) ([]bson.M, error) {
	cursor, err := m.database.Collection(collection).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err = cursor.All(ctx, &indexes); err != nil {
		return nil, err
	}

	return indexes, nil
}

// MongoCollection 实现 Collection 接口
type MongoCollection struct {
	collection *mongo.Collection
}

// InsertOne 插入单个文档
func (c *MongoCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return c.collection.InsertOne(ctx, document)
}

// InsertMany 插入多个文档
func (c *MongoCollection) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	return c.collection.InsertMany(ctx, documents)
}

// FindOne 查找单个文档
func (c *MongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	return c.collection.FindOne(ctx, filter)
}

// Find 查找多个文档
func (c *MongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return c.collection.Find(ctx, filter, opts...)
}

// FindOneAndUpdate 查找并更新单个文档
func (c *MongoCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	return c.collection.FindOneAndUpdate(ctx, filter, update, opts...)
}

// FindOneAndReplace 查找并替换单个文档
func (c *MongoCollection) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	return c.collection.FindOneAndReplace(ctx, filter, replacement, opts...)
}

// FindOneAndDelete 查找并删除单个文档
func (c *MongoCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	return c.collection.FindOneAndDelete(ctx, filter, opts...)
}

// UpdateOne 更新单个文档
func (c *MongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.collection.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany 更新多个文档
func (c *MongoCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.collection.UpdateMany(ctx, filter, update, opts...)
}

// ReplaceOne 替换单个文档
func (c *MongoCollection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	return c.collection.ReplaceOne(ctx, filter, replacement, opts...)
}

// DeleteOne 删除单个文档
func (c *MongoCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(ctx, filter, opts...)
}

// DeleteMany 删除多个文档
func (c *MongoCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteMany(ctx, filter, opts...)
}

// Aggregate 聚合查询
func (c *MongoCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return c.collection.Aggregate(ctx, pipeline, opts...)
}

// CountDocuments 统计文档数量
func (c *MongoCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return c.collection.CountDocuments(ctx, filter, opts...)
}

// EstimatedDocumentCount 估算文档数量
func (c *MongoCollection) EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	return c.collection.EstimatedDocumentCount(ctx, opts...)
}

// Distinct 获取唯一值
func (c *MongoCollection) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	return c.collection.Distinct(ctx, fieldName, filter, opts...)
}

// BulkWrite 批量写入
func (c *MongoCollection) BulkWrite(ctx context.Context, operations []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return c.collection.BulkWrite(ctx, operations, opts...)
}

// Drop 删除集合
func (c *MongoCollection) Drop(ctx context.Context) error {
	return c.collection.Drop(ctx)
}

// CreateIndex 创建索引
func (c *MongoCollection) CreateIndex(ctx context.Context, model mongo.IndexModel) (string, error) {
	return c.collection.Indexes().CreateOne(ctx, model)
}

// CreateIndexes 创建多个索引
func (c *MongoCollection) CreateIndexes(ctx context.Context, models []mongo.IndexModel) ([]string, error) {
	return c.collection.Indexes().CreateMany(ctx, models)
}

// DropIndex 删除索引
func (c *MongoCollection) DropIndex(ctx context.Context, name string) error {
	_, err := c.collection.Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes 列出索引
func (c *MongoCollection) ListIndexes(ctx context.Context) (*mongo.Cursor, error) {
	return c.collection.Indexes().List(ctx)
}
