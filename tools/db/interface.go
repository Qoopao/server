package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database 定义数据库连接接口
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

	// 索引管理
	CreateIndex(ctx context.Context, collection string, model mongo.IndexModel) error
	CreateIndexes(ctx context.Context, collection string, models []mongo.IndexModel) error
	DropIndex(ctx context.Context, collection string, name string) error
	ListIndexes(ctx context.Context, collection string) ([]bson.M, error)
}

// Collection 定义集合操作接口
type Collection interface {
	// 基础CRUD操作
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
	InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error)

	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult

	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error)

	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)

	// 聚合操作
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error)

	// 计数和统计
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error)
	Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error)

	// 批量操作
	BulkWrite(ctx context.Context, operations []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error)

	// 集合管理
	Drop(ctx context.Context) error
	CreateIndex(ctx context.Context, model mongo.IndexModel) (string, error)
	CreateIndexes(ctx context.Context, models []mongo.IndexModel) ([]string, error)
	DropIndex(ctx context.Context, name string) error
	ListIndexes(ctx context.Context) (*mongo.Cursor, error)
}

// Repository 定义通用仓储接口
type Repository[T any] interface {
	// 基础CRUD
	Create(ctx context.Context, entity *T) error
	CreateMany(ctx context.Context, entities []*T) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*T, error)
	GetByFilter(ctx context.Context, filter bson.M) (*T, error)
	GetMany(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*T, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) error
	UpdateByFilter(ctx context.Context, filter bson.M, update bson.M) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	DeleteByFilter(ctx context.Context, filter bson.M) error

	// 分页查询
	GetPage(ctx context.Context, filter bson.M, page, size int64, sort bson.M) ([]*T, int64, error)

	// 计数
	Count(ctx context.Context, filter bson.M) (int64, error)

	// 聚合查询
	Aggregate(ctx context.Context, pipeline []bson.M) ([]bson.M, error)

	// 批量操作
	BulkUpdate(ctx context.Context, updates []BulkUpdate) error
	BulkDelete(ctx context.Context, filters []bson.M) error
}

// BulkUpdate 批量更新结构
type BulkUpdate struct {
	Filter bson.M
	Update bson.M
}

// QueryOptions 查询选项
type QueryOptions struct {
	Skip       *int64
	Limit      *int64
	Sort       bson.M
	Projection bson.M
	Collation  *options.Collation
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
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// BeforeCreate 创建前钩子
func (e *BaseEntity) BeforeCreate() {
	now := time.Now()
	if e.ID.IsZero() {
		e.ID = primitive.NewObjectID()
	}
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
