package application

import (
	"context"

	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/infrastructure/persistence"
)

// SequenceUsecase 序列号应用用例接口。
type SequenceUsecase interface {
	// GetNextSeqInc 获取下一个序列号（简单递增）
	GetNextSeqInc(ctx context.Context, id string) (int64, error)

	// GetNextSeqConsecutive 获取下一个序列号（连续递增策略，可与 Inc 共享实现）
	GetNextSeqConsecutive(ctx context.Context, id string) (int64, error)
}

// sequenceUsecase SequenceUsecase 的实现。
type sequenceUsecase struct {
	repo persistence.SequenceRepository
}

// NewSequenceUsecase 创建 SequenceUsecase 实例。
func NewSequenceUsecase(repo persistence.SequenceRepository) SequenceUsecase {
	return &sequenceUsecase{
		repo: repo,
	}
}

// GetNextSeqInc 获取下一个序列号（简单递增）
func (s *sequenceUsecase) GetNextSeqInc(ctx context.Context, id string) (int64, error) {
	return s.repo.GetNextSeq(ctx, id)
}

// GetNextSeqConsecutive 获取下一个序列号（连续递增）
// 当前实现与 GetNextSeqInc 相同，如后续有“补洞”等策略可在此单独扩展。
func (s *sequenceUsecase) GetNextSeqConsecutive(ctx context.Context, id string) (int64, error) {
	return s.repo.GetNextSeq(ctx, id)
}
