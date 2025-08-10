package kvstore

import (
	"context"
	"errors"
	"time"
)

// KVStore 统一的键值存储接口
type KVStore interface {
	// 基本操作
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 高级操作
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 批量操作
	MGet(ctx context.Context, keys ...string) (map[string][]byte, error)
	MSet(ctx context.Context, data map[string][]byte, ttl time.Duration) error

	// 哈希表操作
	HGet(ctx context.Context, key, field string) ([]byte, error)
	HSet(ctx context.Context, key, field string, value []byte) error
	HGetAll(ctx context.Context, key string) (map[string][]byte, error)

	// 列表操作
	LPush(ctx context.Context, key string, values ...[]byte) (int64, error)
	RPush(ctx context.Context, key string, values ...[]byte) (int64, error)
	LPop(ctx context.Context, key string) ([]byte, error)
	RPop(ctx context.Context, key string) ([]byte, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)
	LRem(ctx context.Context, key string, count int64, value []byte) error

	// 集合操作
	SAdd(ctx context.Context, key string, members ...[]byte) error
	SRem(ctx context.Context, key string, members ...[]byte) error
	SMembers(ctx context.Context, key string) ([][]byte, error)

	// 关闭连接
	Close() error
}

// 错误类型定义
var (
	ErrKVStoreNotInitialized = errors.New("not initialized")
	ErrKeyNotFound           = errors.New("key not found")
	ErrInvalidType           = errors.New("invalid value type")
	ErrTimeout               = errors.New("operation timed out")
)

// Config 存储配置
type Config struct {
	Address  string        // Redis地址: localhost:6379
	Password string        // Redis密码
	DB       int           // Redis数据库
	Timeout  time.Duration // 连接超时
}

func NewKVStore(cfg Config) (KVStore, error) {
	ctx := context.Background()
	instance, err := newRedisStore(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return instance, nil
}
