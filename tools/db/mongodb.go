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
func (m *MongoDB) CreateIndex(ctx context.Context, collection string, model IndexModel) error {
	mongoModel := convertToMongoIndexModel(model)
	_, err := m.database.Collection(collection).Indexes().CreateOne(ctx, mongoModel)
	return err
}

// CreateIndexes 创建多个索引
func (m *MongoDB) CreateIndexes(ctx context.Context, collection string, models []IndexModel) error {
	mongoModels := make([]mongo.IndexModel, 0, len(models))
	for _, mdl := range models {
		mongoModels = append(mongoModels, convertToMongoIndexModel(mdl))
	}
	_, err := m.database.Collection(collection).Indexes().CreateMany(ctx, mongoModels)
	return err
}

// DropIndex 删除索引
func (m *MongoDB) DropIndex(ctx context.Context, collection string, name string) error {
	_, err := m.database.Collection(collection).Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes 列出索引
func (m *MongoDB) ListIndexes(ctx context.Context, collection string) ([]map[string]any, error) {
	cursor, err := m.database.Collection(collection).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []map[string]any
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
func (c *MongoCollection) InsertOne(ctx context.Context, document any) (*InsertOneResult, error) {
	res, err := c.collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}
	return &InsertOneResult{InsertedID: res.InsertedID}, nil
}

// InsertMany 插入多个文档
func (c *MongoCollection) InsertMany(ctx context.Context, documents []any) (*InsertManyResult, error) {
	res, err := c.collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, err
	}
	ids := make([]any, 0, len(res.InsertedIDs))
	ids = append(ids, res.InsertedIDs...)
	return &InsertManyResult{InsertedIDs: ids}, nil
}

// FindOne 查找单个文档
func (c *MongoCollection) FindOne(ctx context.Context, filter any) SingleResult {
	return &mongoSingleResult{sr: c.collection.FindOne(ctx, filter)}
}

// Find 查找多个文档
func (c *MongoCollection) Find(ctx context.Context, filter any, opts ...*FindOptions) (Cursor, error) {
	mopts := convertFindOptions(opts...)
	cur, err := c.collection.Find(ctx, filter, mopts...)
	if err != nil {
		return nil, err
	}
	return &mongoCursor{cur: cur}, nil
}

// FindOneAndUpdate 查找并更新单个文档
func (c *MongoCollection) FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...*FindOneAndUpdateOptions) SingleResult {
	mopts := convertFindOneAndUpdateOptions(opts...)
	return &mongoSingleResult{sr: c.collection.FindOneAndUpdate(ctx, filter, update, mopts...)}
}

// FindOneAndReplace 查找并替换单个文档
func (c *MongoCollection) FindOneAndReplace(ctx context.Context, filter any, replacement any, opts ...*FindOneAndReplaceOptions) SingleResult {
	mopts := convertFindOneAndReplaceOptions(opts...)
	return &mongoSingleResult{sr: c.collection.FindOneAndReplace(ctx, filter, replacement, mopts...)}
}

// FindOneAndDelete 查找并删除单个文档
func (c *MongoCollection) FindOneAndDelete(ctx context.Context, filter any, opts ...*FindOneAndDeleteOptions) SingleResult {
	mopts := convertFindOneAndDeleteOptions(opts...)
	return &mongoSingleResult{sr: c.collection.FindOneAndDelete(ctx, filter, mopts...)}
}

// UpdateOne 更新单个文档
func (c *MongoCollection) UpdateOne(ctx context.Context, filter any, update any, opts ...*UpdateOptions) (*UpdateResult, error) {
	mopts := convertUpdateOptions(opts...)
	res, err := c.collection.UpdateOne(ctx, filter, update, mopts...)
	if err != nil {
		return nil, err
	}
	return &UpdateResult{
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
		UpsertedCount: res.UpsertedCount,
		UpsertedID:    res.UpsertedID,
	}, nil
}

// UpdateMany 更新多个文档
func (c *MongoCollection) UpdateMany(ctx context.Context, filter any, update any, opts ...*UpdateOptions) (*UpdateResult, error) {
	mopts := convertUpdateOptions(opts...)
	res, err := c.collection.UpdateMany(ctx, filter, update, mopts...)
	if err != nil {
		return nil, err
	}
	return &UpdateResult{
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
		UpsertedCount: res.UpsertedCount,
		UpsertedID:    res.UpsertedID,
	}, nil
}

// ReplaceOne 替换单个文档
func (c *MongoCollection) ReplaceOne(ctx context.Context, filter any, replacement any, opts ...*ReplaceOptions) (*UpdateResult, error) {
	mopts := convertReplaceOptions(opts...)
	res, err := c.collection.ReplaceOne(ctx, filter, replacement, mopts...)
	if err != nil {
		return nil, err
	}
	return &UpdateResult{
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
		UpsertedCount: res.UpsertedCount,
		UpsertedID:    res.UpsertedID,
	}, nil
}

// DeleteOne 删除单个文档
func (c *MongoCollection) DeleteOne(ctx context.Context, filter any, opts ...*DeleteOptions) (*DeleteResult, error) {
	mopts := convertDeleteOptions(opts...)
	res, err := c.collection.DeleteOne(ctx, filter, mopts...)
	if err != nil {
		return nil, err
	}
	return &DeleteResult{DeletedCount: res.DeletedCount}, nil
}

// DeleteMany 删除多个文档
func (c *MongoCollection) DeleteMany(ctx context.Context, filter any, opts ...*DeleteOptions) (*DeleteResult, error) {
	mopts := convertDeleteOptions(opts...)
	res, err := c.collection.DeleteMany(ctx, filter, mopts...)
	if err != nil {
		return nil, err
	}
	return &DeleteResult{DeletedCount: res.DeletedCount}, nil
}

// Aggregate 聚合查询
func (c *MongoCollection) Aggregate(ctx context.Context, pipeline any, opts ...*AggregateOptions) (Cursor, error) {
	conv := convertPipelineOrderedDocs(pipeline)
	cur, err := c.collection.Aggregate(ctx, conv)
	if err != nil {
		return nil, err
	}
	return &mongoCursor{cur: cur}, nil
}

// convertPipelineOrderedDocs 将管道中出现的 []KeyValue 转换为 bson.D 以保证顺序
func convertPipelineOrderedDocs(pipeline any) any {
	switch p := pipeline.(type) {
	case []map[string]any:
		out := make([]any, 0, len(p))
		for _, stage := range p {
			out = append(out, convertStageOrderedDocs(stage))
		}
		return out
	case []any:
		out := make([]any, 0, len(p))
		for _, s := range p {
			if m, ok := s.(map[string]any); ok {
				out = append(out, convertStageOrderedDocs(m))
			} else {
				out = append(out, s)
			}
		}
		return out
	default:
		return pipeline
	}
}

func convertStageOrderedDocs(stage map[string]any) map[string]any {
	out := make(map[string]any, len(stage))
	for k, v := range stage {
		switch vv := v.(type) {
		case []KeyValue:
			out[k] = keyValuesToBsonD(vv)
		default:
			out[k] = vv
		}
	}
	return out
}

func keyValuesToBsonD(kvs []KeyValue) bson.D {
	d := make(bson.D, 0, len(kvs))
	for _, kv := range kvs {
		d = append(d, bson.E{Key: kv.Key, Value: kv.Value})
	}
	return d
}

// CountDocuments 统计文档数量
func (c *MongoCollection) CountDocuments(ctx context.Context, filter any, opts ...*CountOptions) (int64, error) {
	return c.collection.CountDocuments(ctx, filter)
}

// EstimatedDocumentCount 估算文档数量
func (c *MongoCollection) EstimatedDocumentCount(ctx context.Context, opts ...*EstimatedDocumentCountOptions) (int64, error) {
	return c.collection.EstimatedDocumentCount(ctx)
}

// Distinct 获取唯一值
func (c *MongoCollection) Distinct(ctx context.Context, fieldName string, filter any, opts ...*DistinctOptions) ([]any, error) {
	vals, err := c.collection.Distinct(ctx, fieldName, filter)
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, len(vals))
	out = append(out, vals...)
	return out, nil
}

// BulkWrite 批量写入
func (c *MongoCollection) BulkWrite(ctx context.Context, operations []WriteOperation, opts ...*BulkWriteOptions) (*BulkWriteResult, error) {
	// 仅支持传入 mongo.WriteModel 的切片
	mongoOps := make([]mongo.WriteModel, 0, len(operations))
	for _, op := range operations {
		m, ok := op.(mongo.WriteModel)
		if !ok {
			return nil, fmt.Errorf("unsupported write operation type: %T", op)
		}
		mongoOps = append(mongoOps, m)
	}
	mopts := convertBulkWriteOptions(opts...)
	res, err := c.collection.BulkWrite(ctx, mongoOps, mopts...)
	if err != nil {
		return nil, err
	}
	return &BulkWriteResult{
		InsertedCount: res.InsertedCount,
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
		DeletedCount:  res.DeletedCount,
		UpsertedCount: res.UpsertedCount,
	}, nil
}

// Drop 删除集合
func (c *MongoCollection) Drop(ctx context.Context) error {
	return c.collection.Drop(ctx)
}

// CreateIndex 创建索引
func (c *MongoCollection) CreateIndex(ctx context.Context, model IndexModel) (string, error) {
	mongoModel := convertToMongoIndexModel(model)
	return c.collection.Indexes().CreateOne(ctx, mongoModel)
}

// CreateIndexes 创建多个索引
func (c *MongoCollection) CreateIndexes(ctx context.Context, models []IndexModel) ([]string, error) {
	mongoModels := make([]mongo.IndexModel, 0, len(models))
	for _, mdl := range models {
		mongoModels = append(mongoModels, convertToMongoIndexModel(mdl))
	}
	return c.collection.Indexes().CreateMany(ctx, mongoModels)
}

// DropIndex 删除索引
func (c *MongoCollection) DropIndex(ctx context.Context, name string) error {
	_, err := c.collection.Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes 列出索引
func (c *MongoCollection) ListIndexes(ctx context.Context) (Cursor, error) {
	cur, err := c.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	return &mongoCursor{cur: cur}, nil
}

// 适配器类型与选项转换

type mongoCursor struct {
	cur *mongo.Cursor
}

func (c *mongoCursor) All(ctx context.Context, results any) error { return c.cur.All(ctx, results) }
func (c *mongoCursor) Next(ctx context.Context) bool              { return c.cur.Next(ctx) }
func (c *mongoCursor) Decode(val any) error                       { return c.cur.Decode(val) }
func (c *mongoCursor) Close(ctx context.Context) error            { return c.cur.Close(ctx) }
func (c *mongoCursor) Err() error                                 { return c.cur.Err() }

type mongoSingleResult struct {
	sr *mongo.SingleResult
}

func (s *mongoSingleResult) Decode(val any) error { return s.sr.Decode(val) }
func (s *mongoSingleResult) Err() error           { return s.sr.Err() }

func convertFindOptions(opts ...*FindOptions) []*options.FindOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.Find()
	if opts[0].Skip != nil {
		o.SetSkip(*opts[0].Skip)
	}
	if opts[0].Limit != nil {
		o.SetLimit(*opts[0].Limit)
	}
	if opts[0].Sort != nil {
		o.SetSort(opts[0].Sort)
	}
	if opts[0].Projection != nil {
		o.SetProjection(opts[0].Projection)
	}
	return []*options.FindOptions{o}
}

func convertFindOneAndUpdateOptions(opts ...*FindOneAndUpdateOptions) []*options.FindOneAndUpdateOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.FindOneAndUpdate()
	if opts[0].ReturnDocument == "after" {
		o.SetReturnDocument(options.After)
	} else if opts[0].ReturnDocument == "before" {
		o.SetReturnDocument(options.Before)
	}
	o.SetUpsert(opts[0].Upsert)
	return []*options.FindOneAndUpdateOptions{o}
}

func convertFindOneAndReplaceOptions(opts ...*FindOneAndReplaceOptions) []*options.FindOneAndReplaceOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.FindOneAndReplace()
	o.SetUpsert(opts[0].Upsert)
	return []*options.FindOneAndReplaceOptions{o}
}

func convertFindOneAndDeleteOptions(opts ...*FindOneAndDeleteOptions) []*options.FindOneAndDeleteOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.FindOneAndDelete()
	return []*options.FindOneAndDeleteOptions{o}
}

func convertUpdateOptions(opts ...*UpdateOptions) []*options.UpdateOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.Update().SetUpsert(opts[0].Upsert)
	return []*options.UpdateOptions{o}
}

func convertReplaceOptions(opts ...*ReplaceOptions) []*options.ReplaceOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.Replace().SetUpsert(opts[0].Upsert)
	return []*options.ReplaceOptions{o}
}

func convertDeleteOptions(opts ...*DeleteOptions) []*options.DeleteOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.Delete()
	return []*options.DeleteOptions{o}
}

func convertBulkWriteOptions(opts ...*BulkWriteOptions) []*options.BulkWriteOptions {
	if len(opts) == 0 || opts[0] == nil {
		return nil
	}
	o := options.BulkWrite().SetOrdered(opts[0].Ordered)
	return []*options.BulkWriteOptions{o}
}

func convertToMongoIndexModel(im IndexModel) mongo.IndexModel {
	var idxOpts *options.IndexOptions
	if im.Options != nil {
		idxOpts = options.Index()
		if im.Options.Name != nil {
			idxOpts.SetName(*im.Options.Name)
		}
		if im.Options.Unique != nil {
			idxOpts.SetUnique(*im.Options.Unique)
		}
		if im.Options.Background != nil {
			idxOpts.SetBackground(*im.Options.Background)
		}
		if im.Options.Sparse != nil {
			idxOpts.SetSparse(*im.Options.Sparse)
		}
	}
	// Keys 现在强制为有序的 []KeyValue，转成 bson.D 以保持顺序
	d := make(bson.D, 0, len(im.Keys))
	for _, kv := range im.Keys {
		d = append(d, bson.E{Key: kv.Key, Value: kv.Value})
	}
	return mongo.IndexModel{Keys: d, Options: idxOpts}
}
