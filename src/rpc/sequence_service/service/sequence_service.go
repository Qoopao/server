package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/storage"
)

// SequenceService 序列号服务接口（逻辑层）
type SequenceService interface {
	// GetNextSeqInc 获取下一个序列号（简单递增）
	GetNextSeqInc(ctx context.Context, id string) (int64, error)

	// GetNextSeqConsecutive 获取下一个序列号（连续递增策略，可与 Inc 共享实现）
	GetNextSeqConsecutive(ctx context.Context, id string) (int64, error)
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

// GetNextSeqInc 获取下一个序列号（简单递增）
func (s *sequenceServiceImpl) GetNextSeqInc(ctx context.Context, id string) (int64, error) {
	return s.storage.GetNextSeq(ctx, id)
}

// GetNextSeqConsecutive 获取下一个序列号（连续递增）
// 当前实现与 GetNextSeqInc 相同，如后续有“补洞”等策略可在此单独扩展。
func (s *sequenceServiceImpl) GetNextSeqConsecutive(ctx context.Context, id string) (int64, error) {
	return s.storage.GetNextSeq(ctx, id)
}
