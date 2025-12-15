package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// SequenceStorage 序列号存储接口（存储层）
type SequenceStorage interface {
	// GetNextSeq 获取下一个序列号
	GetNextSeq(ctx context.Context, conversationID string) (int64, error)

	// BatchGetNextSeq 批量获取序列号
	BatchGetNextSeq(ctx context.Context, conversationIDs []string) (map[string]int64, error)

	// GetMaxSeq 获取当前最大序列号
	GetMaxSeq(ctx context.Context, conversationID string) (int64, error)
}

// sequenceStorageImpl SequenceStorage的实现，使用Redis
type sequenceStorageImpl struct {
	client *redis.Client
}

// NewSequenceStorage 创建SequenceStorage实例
func NewSequenceStorage(client *redis.Client) SequenceStorage {
	return &sequenceStorageImpl{
		client: client,
	}
}

// keyForSeq 生成seq的Redis key
func keyForSeq(conversationID string) string {
	return fmt.Sprintf("seq:conv:%s", conversationID)
}

// GetNextSeq 获取下一个序列号（使用Redis INCR）
func (s *sequenceStorageImpl) GetNextSeq(ctx context.Context, conversationID string) (int64, error) {
	if conversationID == "" {
		return 0, fmt.Errorf("conversationID is empty")
	}

	key := keyForSeq(conversationID)

	// 使用Redis INCR命令，原子递增并返回新值
	seq, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment seq for conversation %s: %w", conversationID, err)
	}

	return seq, nil
}

// BatchGetNextSeq 批量获取序列号
func (s *sequenceStorageImpl) BatchGetNextSeq(ctx context.Context, conversationIDs []string) (map[string]int64, error) {
	result := make(map[string]int64, len(conversationIDs))

	for _, convID := range conversationIDs {
		seq, err := s.GetNextSeq(ctx, convID)
		if err != nil {
			// 单个失败不影响其他，记录错误但继续处理
			continue
		}
		result[convID] = seq
	}

	return result, nil
}

// GetMaxSeq 获取当前最大序列号（获取当前值，不递增）
func (s *sequenceStorageImpl) GetMaxSeq(ctx context.Context, conversationID string) (int64, error) {
	if conversationID == "" {
		return 0, fmt.Errorf("conversationID is empty")
	}

	key := keyForSeq(conversationID)

	// 获取当前值，如果不存在返回0
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// key不存在时返回0
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get max seq for conversation %s: %w", conversationID, err)
	}

	// 解析为int64
	seq, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse seq value for conversation %s: %w", conversationID, err)
	}

	return seq, nil
}
