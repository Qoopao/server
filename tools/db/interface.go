package db

import (
	"context"
	"time"
)

// Database 定义数据库连接接口（与具体实现解耦）
type Database interface {
	// 连接管理
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error

	// 数据库操作
	GetDatabase(name string) Database
	GetCollection(name string) Collection

	// 事务支持
	WithTransaction(ctx context.Context, fn func(context.Context) error) error

	// 索引管理（使用与实现无关的模型与选项）
	CreateIndex(ctx context.Context, collection string, model IndexModel) error
	CreateIndexes(ctx context.Context, collection string, models []IndexModel) error
	DropIndex(ctx context.Context, collection string, name string) error
	ListIndexes(ctx context.Context, collection string) ([]map[string]any, error)
}

// Collection 定义集合操作接口（与具体实现解耦）
type Collection interface {
	// 基础CRUD操作
	InsertOne(ctx context.Context, document any) (*InsertOneResult, error)
	InsertMany(ctx context.Context, documents []any) (*InsertManyResult, error)

	FindOne(ctx context.Context, filter any) SingleResult
	Find(ctx context.Context, filter any, opts ...*FindOptions) (Cursor, error)
	FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...*FindOneAndUpdateOptions) SingleResult
	FindOneAndReplace(ctx context.Context, filter any, replacement any, opts ...*FindOneAndReplaceOptions) SingleResult
	FindOneAndDelete(ctx context.Context, filter any, opts ...*FindOneAndDeleteOptions) SingleResult

	UpdateOne(ctx context.Context, filter any, update any, opts ...*UpdateOptions) (*UpdateResult, error)
	UpdateMany(ctx context.Context, filter any, update any, opts ...*UpdateOptions) (*UpdateResult, error)
	ReplaceOne(ctx context.Context, filter any, replacement any, opts ...*ReplaceOptions) (*UpdateResult, error)

	DeleteOne(ctx context.Context, filter any, opts ...*DeleteOptions) (*DeleteResult, error)
	DeleteMany(ctx context.Context, filter any, opts ...*DeleteOptions) (*DeleteResult, error)

	// 聚合操作
	Aggregate(ctx context.Context, pipeline any, opts ...*AggregateOptions) (Cursor, error)

	// 计数和统计
	CountDocuments(ctx context.Context, filter any, opts ...*CountOptions) (int64, error)
	EstimatedDocumentCount(ctx context.Context, opts ...*EstimatedDocumentCountOptions) (int64, error)
	Distinct(ctx context.Context, fieldName string, filter any, opts ...*DistinctOptions) ([]any, error)

	// 批量操作
	BulkWrite(ctx context.Context, operations []WriteOperation, opts ...*BulkWriteOptions) (*BulkWriteResult, error)

	// 集合管理
	Drop(ctx context.Context) error
	CreateIndex(ctx context.Context, model IndexModel) (string, error)
	CreateIndexes(ctx context.Context, models []IndexModel) ([]string, error)
	DropIndex(ctx context.Context, name string) error
	ListIndexes(ctx context.Context) (Cursor, error)
}

// Repository 定义通用仓储接口
type Repository[T any] interface {
	// 基础CRUD
	Create(ctx context.Context, entity *T) error
	CreateMany(ctx context.Context, entities []*T) error
	GetByID(ctx context.Context, id any) (*T, error)
	GetByFilter(ctx context.Context, filter map[string]any) (*T, error)
	GetMany(ctx context.Context, filter map[string]any, opts ...*FindOptions) ([]*T, error)
	UpdateByID(ctx context.Context, id any, update map[string]any) error
	UpdateByFilter(ctx context.Context, filter map[string]any, update map[string]any) error
	DeleteByID(ctx context.Context, id any) error
	DeleteByFilter(ctx context.Context, filter map[string]any) error

	// 分页查询
	GetPage(ctx context.Context, filter map[string]any, page, size int64, sort map[string]any) ([]*T, int64, error)

	// 计数
	Count(ctx context.Context, filter map[string]any) (int64, error)

	// 聚合查询
	Aggregate(ctx context.Context, pipeline []map[string]any) ([]map[string]any, error)

	// 批量操作
	BulkUpdate(ctx context.Context, updates []BulkUpdate) error
	BulkDelete(ctx context.Context, filters []map[string]any) error
}

// BulkUpdate 批量更新结构
type BulkUpdate struct {
	Filter map[string]any
	Update map[string]any
}

// QueryOptions 查询选项
type QueryOptions struct {
	Skip       *int64
	Limit      *int64
	Sort       map[string]any
	Projection map[string]any
	Collation  *Collation
}

// 通用结果与选项类型（与具体实现无关）

// InsertOneResult 表示插入单条记录的结果
type InsertOneResult struct {
	InsertedID any
}

// InsertManyResult 表示插入多条记录的结果
type InsertManyResult struct {
	InsertedIDs []any
}

// UpdateResult 表示更新操作的结果
type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    any
}

// DeleteResult 表示删除操作的结果
type DeleteResult struct {
	DeletedCount int64
}

// BulkWriteResult 表示批量写入的结果
type BulkWriteResult struct {
	InsertedCount int64
	MatchedCount  int64
	ModifiedCount int64
	DeletedCount  int64
	UpsertedCount int64
}

// Cursor 抽象游标
type Cursor interface {
	All(ctx context.Context, results any) error
	Next(ctx context.Context) bool
	Decode(val any) error
	Close(ctx context.Context) error
	Err() error
}

// SingleResult 抽象单条结果
type SingleResult interface {
	Decode(val any) error
	Err() error
}

// 通用选项类型（子集，用于跨存储抽象）。具体实现可忽略不支持字段。
type FindOptions struct {
	Skip       *int64
	Limit      *int64
	Sort       map[string]any
	Projection map[string]any
}

type FindOneAndUpdateOptions struct {
	ReturnDocument string // "before" | "after"
	Upsert         bool
}

type FindOneAndReplaceOptions struct {
	Upsert bool
}

type FindOneAndDeleteOptions struct{}

type UpdateOptions struct {
	Upsert bool
}

type ReplaceOptions struct {
	Upsert bool
}

type DeleteOptions struct{}

type AggregateOptions struct{}

type CountOptions struct{}

type EstimatedDocumentCountOptions struct{}

type DistinctOptions struct{}

type BulkWriteOptions struct {
	Ordered bool
}

// 批量写操作抽象（具体实现可通过类型断言区分）
type WriteOperation interface{}

// 索引相关抽象
type IndexModel struct {
	Keys    any           // 通常为 map 或数组文档
	Options *IndexOptions // 选项
}

type IndexOptions struct {
	Name       *string
	Unique     *bool
	Background *bool
	Sparse     *bool
}

// Collation 抽象
type Collation struct {
	Locale          string
	CaseLevel       *bool
	CaseFirst       *string
	Strength        *int
	NumericOrdering *bool
	Alternate       *string
	MaxVariable     *string
	Backwards       *bool
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	URI            string
	Database       string
	MaxPoolSize    uint64
	MinPoolSize    uint64
	MaxIdleTime    time.Duration
	ConnectTimeout time.Duration
	SocketTimeout  time.Duration
	ServerTimeout  time.Duration
	AuthSource     string
	Username       string
	Password       string
	SSL            bool
	ReplicaSet     string
}

// DefaultConnectionConfig 返回默认连接配置
func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "roc_im",
		MaxPoolSize:    100,
		MinPoolSize:    10,
		MaxIdleTime:    30 * time.Minute,
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  30 * time.Second,
		ServerTimeout:  30 * time.Second,
		SSL:            false,
	}
}

// BaseEntity 基础实体结构，包含通用字段
type BaseEntity struct {
	ID        any        `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// BeforeCreate 创建前钩子
func (e *BaseEntity) BeforeCreate() {
	now := time.Now()
	// ID 的生成由具体实现或上层业务负责
	e.CreatedAt = now
	e.UpdatedAt = now
}

// BeforeUpdate 更新前钩子
func (e *BaseEntity) BeforeUpdate() {
	e.UpdatedAt = time.Now()
}

// SoftDelete 软删除
func (e *BaseEntity) SoftDelete() {
	now := time.Now()
	e.DeletedAt = &now
	e.UpdatedAt = now
}

// IsDeleted 检查是否已删除
func (e *BaseEntity) IsDeleted() bool {
	return e.DeletedAt != nil
}
