package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/storage"
)

// SequenceService 序列号服务接口（逻辑层）
type SequenceService interface {
	// GetNextSeq 获取下一个序列号
	GetNextSeq(ctx context.Context, conversationID string) (int64, error)

	// BatchGetNextSeq 批量获取序列号
	BatchGetNextSeq(ctx context.Context, conversationIDs []string) (map[string]int64, error)

	// GetMaxSeq 获取当前最大序列号
	GetMaxSeq(ctx context.Context, conversationID string) (int64, error)
}

// sequenceServiceImpl SequenceService的实现
type sequenceServiceImpl struct {
	storage storage.SequenceStorage
}

// NewSequenceService 创建SequenceService实例
func NewSequenceService(storage storage.SequenceStorage) SequenceService {
	return &sequenceServiceImpl{
		storage: storage,
	}
}

// GetNextSeq 获取下一个序列号
func (s *sequenceServiceImpl) GetNextSeq(ctx context.Context, conversationID string) (int64, error) {
	return s.storage.GetNextSeq(ctx, conversationID)
}

// BatchGetNextSeq 批量获取序列号
func (s *sequenceServiceImpl) BatchGetNextSeq(ctx context.Context, conversationIDs []string) (map[string]int64, error) {
	return s.storage.BatchGetNextSeq(ctx, conversationIDs)
}

// GetMaxSeq 获取当前最大序列号
func (s *sequenceServiceImpl) GetMaxSeq(ctx context.Context, conversationID string) (int64, error) {
	return s.storage.GetMaxSeq(ctx, conversationID)
}
