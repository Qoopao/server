package sequence

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	sequenceservice "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// InMemorySequenceService 是一个仅用于本地测试的简易实现，不依赖 Redis / Etcd。
type InMemorySequenceService struct {
	mu   sync.Mutex
	data map[string]int64
}

func NewInMemorySequenceService() *InMemorySequenceService {
	return &InMemorySequenceService{
		data: make(map[string]int64),
	}
}

func (s *InMemorySequenceService) GetNextSeq(ctx context.Context, req *sequencepb.GetNextSeqRequest) (*sequencepb.GetNextSeqResponse, error) {
	if req == nil || req.ConversationId == "" {
		return &sequencepb.GetNextSeqResponse{
			ErrorCode: "InvalidArgument",
			ErrorMsg:  "conversation_id is empty",
		}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	seq := s.data[req.ConversationId] + 1
	s.data[req.ConversationId] = seq

	return &sequencepb.GetNextSeqResponse{
		Seq: seq,
	}, nil
}

func (s *InMemorySequenceService) BatchGetNextSeq(ctx context.Context, req *sequencepb.BatchGetNextSeqRequest) (*sequencepb.BatchGetNextSeqResponse, error) {
	res := &sequencepb.BatchGetNextSeqResponse{}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cid := range req.GetConversationIds() {
		if cid == "" {
			res.Results = append(res.Results, &sequencepb.SeqResult{
				ErrorCode: "InvalidArgument",
				ErrorMsg:  "conversation_id is empty",
			})
			continue
		}
		seq := s.data[cid] + 1
		s.data[cid] = seq
		res.Results = append(res.Results, &sequencepb.SeqResult{
			ConversationId: cid,
			Seq:            seq,
		})
	}

	return res, nil
}

func (s *InMemorySequenceService) GetMaxSeq(ctx context.Context, req *sequencepb.GetMaxSeqRequest) (*sequencepb.GetMaxSeqResponse, error) {
	if req == nil || req.ConversationId == "" {
		return &sequencepb.GetMaxSeqResponse{
			ErrorCode: "InvalidArgument",
			ErrorMsg:  "conversation_id is empty",
		}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	seq := s.data[req.ConversationId]
	return &sequencepb.GetMaxSeqResponse{
		MaxSeq: seq,
	}, nil
}

// StartSimpleSequenceServer 在本地启动一个简易的 sequence 服务，仅用于测试。
// host 通常可以传 "127.0.0.1"，port 可以与客户端保持一致（如 10300）。
func StartSimpleSequenceServer(host string, port int) {
	addr := &net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	svr := sequenceservice.NewServer(
		NewInMemorySequenceService(),
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.SequenceServiceName}),
	)

	go func() {
		log.Printf("[sequence-test] simple sequence server listening on %s:%d", host, port)
		if err := svr.Run(); err != nil {
			log.Fatalf("[sequence-test] simple sequence server stopped with error: %v", err)
		}
	}()

	// 给服务器一个短暂时间完成启动，避免客户端立即连接时报错
	time.Sleep(200 * time.Millisecond)
}

