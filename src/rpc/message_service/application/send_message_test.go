package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cloudwego/kitex/client/callopt"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	sequenceclient "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	persistence "github.com/rhp-QE/roc-im-server/src/rpc/message_service/infrastructure/persistence"
)

func TestBatchSendMessagePersistsFactBeforePublish(t *testing.T) {
	storage := newFakeMessageRepository()
	svc := &messageServiceImpl{
		messageRepo: storage,
		serviceCtx:  &fakeServiceContext{seqClient: &fakeSequenceClient{next: 100}},
	}

	resp, err := svc.BatchSendMessage(context.Background(), &sdkws.BatchSendMessageRequest{
		Msgs: []*sdkws.MessageData{{
			SendID:     "u1",
			RecvID:     "u2",
			ConvID:     "single_u1_u2",
			CMessaegID: "client-1",
			ConvType:   1,
		}},
	})
	if err != nil {
		t.Fatalf("BatchSendMessage returned error: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].ErrorCode != "" {
		t.Fatalf("unexpected result: %+v", resp.Results)
	}
	got := resp.Results[0].Msg
	if got.GetSMessageID() == "" || got.GetSeq() != 100 {
		t.Fatalf("message fact was not assigned correctly: %+v", got)
	}
	if !reflect.DeepEqual(storage.ops, []string{"save", "publish"}) {
		t.Fatalf("unexpected operation order: %v", storage.ops)
	}
}

func TestBatchSendMessageMarksMQPublishFailure(t *testing.T) {
	storage := newFakeMessageRepository()
	storage.publishErr = errors.New("mq down")
	svc := &messageServiceImpl{
		messageRepo: storage,
		serviceCtx:  &fakeServiceContext{seqClient: &fakeSequenceClient{next: 7}},
	}

	resp, err := svc.BatchSendMessage(context.Background(), &sdkws.BatchSendMessageRequest{
		Msgs: []*sdkws.MessageData{{
			SendID:     "u1",
			RecvID:     "u2",
			ConvID:     "single_u1_u2",
			CMessaegID: "client-2",
			ConvType:   1,
		}},
	})
	if err != nil {
		t.Fatalf("BatchSendMessage returned error: %v", err)
	}
	if got := resp.Results[0].GetErrorCode(); got != "MQ_PUBLISH_FAILED" {
		t.Fatalf("expected MQ_PUBLISH_FAILED, got %q", got)
	}
	if len(storage.saved) != 1 {
		t.Fatalf("message fact should still be persisted, saved=%d", len(storage.saved))
	}
}

func TestBatchSendMessageReusesPersistedClientMessage(t *testing.T) {
	storage := newFakeMessageRepository()
	existing := &sdkws.MessageData{
		SendID:     "u1",
		RecvID:     "u2",
		ConvID:     "single_u1_u2",
		CMessaegID: "client-3",
		SMessageID: "server-existing",
		Seq:        42,
		ConvType:   1,
	}
	storage.seed(existing)
	seqClient := &fakeSequenceClient{err: errors.New("sequence should not be called for idempotent retry")}
	svc := &messageServiceImpl{
		messageRepo: storage,
		serviceCtx:  &fakeServiceContext{seqClient: seqClient},
	}

	resp, err := svc.BatchSendMessage(context.Background(), &sdkws.BatchSendMessageRequest{
		Msgs: []*sdkws.MessageData{{
			SendID:     "u1",
			RecvID:     "u2",
			ConvID:     "single_u1_u2",
			CMessaegID: "client-3",
			ConvType:   1,
		}},
	})
	if err != nil {
		t.Fatalf("BatchSendMessage returned error: %v", err)
	}
	got := resp.Results[0].Msg
	if got.GetSMessageID() != existing.SMessageID || got.GetSeq() != existing.Seq {
		t.Fatalf("expected existing fact, got %+v", got)
	}
	if seqClient.consecutiveCalls != 0 {
		t.Fatalf("sequence-service should not be called, calls=%d", seqClient.consecutiveCalls)
	}
	if len(storage.published) != 1 || storage.published[0].GetSMessageID() != existing.SMessageID {
		t.Fatalf("expected existing fact to be republished once, published=%+v", storage.published)
	}
}

type fakeServiceContext struct {
	seqClient sequenceclient.Client
	seqErr    error
}

func (f *fakeServiceContext) GetSequenceServiceClient(context.Context) (sequenceclient.Client, error) {
	return f.seqClient, f.seqErr
}

func (f *fakeServiceContext) GetMQProducer() mq.Producer {
	return nil
}

func (f *fakeServiceContext) GetStorage() foundationstorage.Storage {
	return nil
}

func (f *fakeServiceContext) Close() error {
	return nil
}

type fakeSequenceClient struct {
	next             int64
	err              error
	consecutiveCalls int
}

func (f *fakeSequenceClient) GetNextSeqInc(context.Context, *sequencepb.GetNextSeqIncRequest, ...callopt.Option) (*sequencepb.GetNextSeqResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeSequenceClient) GetNextSeqConsecutive(context.Context, *sequencepb.GetNextSeqConsecutiveRequest, ...callopt.Option) (*sequencepb.GetNextSeqResponse, error) {
	f.consecutiveCalls++
	if f.err != nil {
		return nil, f.err
	}
	if f.next == 0 {
		f.next = 1
	}
	seq := f.next
	f.next++
	return &sequencepb.GetNextSeqResponse{Seq: seq}, nil
}

type fakeMessageRepository struct {
	facts      map[string]*sdkws.MessageData
	byClient   map[string]*sdkws.MessageData
	saved      []*sdkws.MessageData
	published  []*sdkws.MessageData
	publishErr error
	ops        []string
}

func newFakeMessageRepository() *fakeMessageRepository {
	return &fakeMessageRepository{
		facts:    make(map[string]*sdkws.MessageData),
		byClient: make(map[string]*sdkws.MessageData),
	}
}

func (f *fakeMessageRepository) seed(msg *sdkws.MessageData) {
	cloned := cloneMessage(msg)
	f.facts[msg.SMessageID] = cloned
	f.byClient[clientKey(msg.SendID, msg.CMessaegID)] = cloned
}

func (f *fakeMessageRepository) SaveMessageFact(_ context.Context, msg *sdkws.MessageData) (*sdkws.MessageData, bool, error) {
	if existing := f.facts[msg.SMessageID]; existing != nil {
		return cloneMessage(existing), true, nil
	}
	cloned := cloneMessage(msg)
	f.facts[msg.SMessageID] = cloned
	f.byClient[clientKey(msg.SendID, msg.CMessaegID)] = cloned
	f.saved = append(f.saved, cloneMessage(msg))
	f.ops = append(f.ops, "save")
	return cloneMessage(msg), false, nil
}

func (f *fakeMessageRepository) FindMessageByClientMsgID(_ context.Context, sendID string, clientMsgID string) (*sdkws.MessageData, bool, error) {
	if msg := f.byClient[clientKey(sendID, clientMsgID)]; msg != nil {
		return cloneMessage(msg), true, nil
	}
	return nil, false, nil
}

func (f *fakeMessageRepository) PublishMessages(_ context.Context, msgs []*sdkws.MessageData) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	for _, msg := range msgs {
		f.published = append(f.published, cloneMessage(msg))
	}
	f.ops = append(f.ops, "publish")
	return nil
}

func (f *fakeMessageRepository) FetchConvMessageListWithRange(context.Context, string, int64, int64) ([]*sdkws.MessageData, error) {
	return nil, nil
}

func (f *fakeMessageRepository) FetchConvLatestMessageList(context.Context, string, int64) ([]*sdkws.MessageData, error) {
	return nil, nil
}

func (f *fakeMessageRepository) BatchGetMessages(context.Context, []string, string) (map[string]*sdkws.MessageData, error) {
	return nil, nil
}

var _ persistence.MessageRepository = (*fakeMessageRepository)(nil)

func clientKey(sendID string, clientMsgID string) string {
	return sendID + "\x00" + clientMsgID
}

func cloneMessage(msg *sdkws.MessageData) *sdkws.MessageData {
	if msg == nil {
		return nil
	}
	cloned := *msg
	return &cloned
}
