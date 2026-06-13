package persistence

import (
	"context"
	"fmt"

	deps "github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/infrastructure/deps"
)

// SequenceRepository 封装 Redis 中的会话序列号原子递增。
type SequenceRepository interface {
	// GetNextSeq 获取下一个序列号
	GetNextSeq(ctx context.Context, id string) (int64, error)
}

// sequenceRepository 使用 Redis INCR 保证单 key 序列原子递增。
type sequenceRepository struct {
	serviceCtx deps.ServiceContext
}

// NewSequenceRepository 创建 SequenceRepository 实例。
func NewSequenceRepository(serviceCtx deps.ServiceContext) SequenceRepository {
	return &sequenceRepository{
		serviceCtx: serviceCtx,
	}
}

// keyForSeq 生成seq的Redis key
func keyForSeq(id string) string {
	return fmt.Sprintf("seq:%s", id)
}

// GetNextSeq 获取下一个序列号（使用Redis INCR）
func (s *sequenceRepository) GetNextSeq(ctx context.Context, id string) (int64, error) {
	if id == "" {
		return 0, fmt.Errorf("id is empty")
	}
	client := s.serviceCtx.GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	key := keyForSeq(id)

	// 使用Redis INCR命令，原子递增并返回新值
	seq, err := client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment seq for id %s: %w", id, err)
	}

	return seq, nil
}
