package storage

import (
	"context"
	"fmt"

	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/service_context"
)

// SequenceStorage 序列号存储接口（存储层）
type SequenceStorage interface {
	// GetNextSeq 获取下一个序列号
	GetNextSeq(ctx context.Context, id string) (int64, error)
}

// sequenceStorageImpl SequenceStorage的实现，使用Redis（通过 ServiceContext 管理）
type sequenceStorageImpl struct {
	serviceCtx servicecontext.ServiceContext
}

// NewSequenceStorage 创建SequenceStorage实例
func NewSequenceStorage(serviceCtx servicecontext.ServiceContext) SequenceStorage {
	return &sequenceStorageImpl{
		serviceCtx: serviceCtx,
	}
}

// keyForSeq 生成seq的Redis key
func keyForSeq(id string) string {
	return fmt.Sprintf("seq:%s", id)
}

// GetNextSeq 获取下一个序列号（使用Redis INCR）
func (s *sequenceStorageImpl) GetNextSeq(ctx context.Context, id string) (int64, error) {
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
