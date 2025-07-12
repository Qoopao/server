// redis_store.go
package kvstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisStore 实现基于Redis的KV存储
type redisStore struct {
	client *redis.Client
}

// newRedisStore 创建Redis存储实例
func newRedisStore(ctx context.Context, config Config) (*redisStore, error) {
	// 创建Redis客户端选项
	opt := &redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	}

	// 设置超时（如果配置）
	if config.Timeout > 0 {
		opt.DialTimeout = config.Timeout
		opt.ReadTimeout = config.Timeout
		opt.WriteTimeout = config.Timeout
	}

	// 创建客户端
	client := redis.NewClient(opt)

	// 测试连接
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisStore{client: client}, nil
}

// 基础操作实现

func (r *redisStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return val, nil
}

func (r *redisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisStore) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// 高级操作实现

func (r *redisStore) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisStore) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

func (r *redisStore) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, ttl).Result()
}

func (r *redisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	dur, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Redis返回-1表示没有设置过期时间，-2表示键不存在
	if dur == -2 {
		return 0, ErrKeyNotFound
	}
	if dur == -1 {
		return 0, nil // 永不过期
	}

	return dur, nil
}

// 批量操作实现

func (r *redisStore) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return map[string][]byte{}, nil
	}

	// 执行MGET命令
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	// 构建结果映射
	result := make(map[string][]byte, len(keys))
	for i, key := range keys {
		val := values[i]
		if val == nil {
			// 键不存在
			continue
		}

		// 类型断言
		switch v := val.(type) {
		case string:
			result[key] = []byte(v)
		case []byte:
			result[key] = v
		default:
			// Redis返回其他类型，转换为字节切片
			result[key] = []byte(fmt.Sprintf("%v", v))
		}
	}

	return result, nil
}

func (r *redisStore) MSet(ctx context.Context, data map[string][]byte, ttl time.Duration) error {
	if len(data) == 0 {
		return nil
	}

	// 使用管道提高性能
	pipe := r.client.Pipeline()

	// 准备所有设置操作
	for key, value := range data {
		pipe.Set(ctx, key, value, ttl)
	}

	// 执行所有命令
	_, err := pipe.Exec(ctx)
	return err
}

// 哈希表操作实现

func (r *redisStore) HGet(ctx context.Context, key, field string) ([]byte, error) {
	val, err := r.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return val, nil
}

func (r *redisStore) HSet(ctx context.Context, key, field string, value []byte) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

func (r *redisStore) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 转换结果为字节切片
	data := make(map[string][]byte, len(result))
	for k, v := range result {
		data[k] = []byte(v)
	}

	return data, nil
}

// 列表操作实现

func (r *redisStore) LPush(ctx context.Context, key string, values ...[]byte) error {
	// 转换为接口切片
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}

	return r.client.LPush(ctx, key, vals...).Err()
}

func (r *redisStore) RPop(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.RPop(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return val, nil
}

// 集合操作实现

func (r *redisStore) SAdd(ctx context.Context, key string, members ...[]byte) error {
	// 转换为接口切片
	mems := make([]interface{}, len(members))
	for i, m := range members {
		mems[i] = m
	}

	return r.client.SAdd(ctx, key, mems...).Err()
}

func (r *redisStore) SRem(ctx context.Context, key string, members ...[]byte) error {
	// 转换为接口切片
	mems := make([]interface{}, len(members))
	for i, m := range members {
		mems[i] = m
	}

	return r.client.SRem(ctx, key, mems...).Err()
}

func (r *redisStore) SMembers(ctx context.Context, key string) ([][]byte, error) {
	vals, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 转换结果为字节切片
	result := make([][]byte, len(vals))
	for i, v := range vals {
		result[i] = []byte(v)
	}

	return result, nil
}

// 关闭连接
func (r *redisStore) Close() error {
	return r.client.Close()
}

// 扩展功能（非接口必需但有用的功能）

// ScanKeys 扫描匹配模式的键
func (r *redisStore) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	var err error

	for {
		// 使用SCAN命令迭代获取键
		var result []string
		result, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, result...)

		// 迭代完成
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// Publish 发布消息到频道
func (r *redisStore) Publish(ctx context.Context, channel string, message []byte) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (r *redisStore) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// Pipeline 创建管道
func (r *redisStore) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline 创建事务管道
func (r *redisStore) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// Stats 获取连接池统计信息
func (r *redisStore) Stats() *redis.PoolStats {
	stats := r.client.PoolStats()
	return stats
}

// Info 获取Redis服务器信息
func (r *redisStore) Info(ctx context.Context) (map[string]string, error) {
	result, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	lines := splitLines(result)
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}

		if idx := indexRune(line, ':'); idx != -1 {
			key := line[:idx]
			value := line[idx+1:]
			info[key] = value
		}
	}

	return info, nil
}

// helper functions
func splitLines(s string) []string {
	var lines []string
	last := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[last:i])
			last = i + 1
		}
	}
	if last < len(s) {
		lines = append(lines, s[last:])
	}
	return lines
}

func indexRune(s string, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}
