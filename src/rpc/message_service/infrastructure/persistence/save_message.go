package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// SaveMessageFact writes the durable message fact before any asynchronous
// projection or push work runs. Duplicate server_msg_id writes return the
// existing fact so client retries converge on the original seq and msg_id.
func (s *messageRepository) SaveMessageFact(ctx context.Context, msg *sdkws.MessageData) (*sdkws.MessageData, bool, error) {
	if msg == nil {
		return nil, false, errors.New("message is nil")
	}
	if msg.SMessageID == "" {
		return nil, false, errors.New("server message id is required")
	}
	if msg.ConvID == "" {
		return nil, false, errors.New("conv id is required")
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, false, errors.New("storage is nil")
	}

	doc, err := buildMessageDocument(msg)
	if err != nil {
		return nil, false, err
	}

	if _, err = store.InsertOne(ctx, orm.CollectionMessages, doc); err != nil {
		if errors.Is(err, foundationstorage.ErrDuplicateKey) {
			existing, found, findErr := s.getMessageByServerID(ctx, msg.SMessageID)
			if findErr != nil {
				return nil, false, findErr
			}
			if !found {
				return nil, false, fmt.Errorf("message duplicate key but existing fact not found: %s", msg.SMessageID)
			}
			klog.CtxInfof(ctx, "[MessageRepository] message fact already exists",
				"server_msg_id", existing.SMessageID,
				"client_msg_id", existing.CMessaegID,
				"conv_id", existing.ConvID,
				"seq", existing.Seq)
			return existing, true, nil
		}
		klog.CtxErrorf(ctx, "[MessageRepository] save message fact failed",
			"server_msg_id", msg.SMessageID,
			"client_msg_id", msg.CMessaegID,
			"conv_id", msg.ConvID,
			"seq", msg.Seq,
			"error", err.Error())
		return nil, false, err
	}

	klog.CtxInfof(ctx, "[MessageRepository] message fact saved",
		"server_msg_id", msg.SMessageID,
		"client_msg_id", msg.CMessaegID,
		"conv_id", msg.ConvID,
		"seq", msg.Seq)
	return msg, false, nil
}

// FindMessageByClientMsgID finds the existing durable fact for client retry
// idempotency. Missing client_msg_id means the caller cannot be deduplicated.
func (s *messageRepository) FindMessageByClientMsgID(ctx context.Context, sendID string, clientMsgID string) (*sdkws.MessageData, bool, error) {
	if sendID == "" || clientMsgID == "" {
		return nil, false, nil
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, false, errors.New("storage is nil")
	}

	var doc orm.MessageDocument
	err := store.FindOne(ctx, orm.CollectionMessages, bson.M{
		"send_id":       sendID,
		"client_msg_id": clientMsgID,
	}, &doc)
	if err != nil {
		if errors.Is(err, foundationstorage.ErrNotFound) {
			return nil, false, nil
		}
		klog.CtxErrorf(ctx, "[MessageRepository] find message by client msg id failed",
			"send_id", sendID,
			"client_msg_id", clientMsgID,
			"error", err.Error())
		return nil, false, err
	}

	msg, err := messageFromDocument(ctx, &doc)
	if err != nil {
		return nil, false, err
	}
	return msg, true, nil
}

func (s *messageRepository) getMessageByServerID(ctx context.Context, serverMsgID string) (*sdkws.MessageData, bool, error) {
	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, false, errors.New("storage is nil")
	}

	var doc orm.MessageDocument
	err := store.FindOne(ctx, orm.CollectionMessages, bson.M{"_id": serverMsgID}, &doc)
	if err != nil {
		if errors.Is(err, foundationstorage.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	msg, err := messageFromDocument(ctx, &doc)
	if err != nil {
		return nil, false, err
	}
	return msg, true, nil
}

func buildMessageDocument(msg *sdkws.MessageData) (*orm.MessageDocument, error) {
	data, err := msg.Marshal(nil)
	if err != nil {
		return nil, fmt.Errorf("marshal message fact failed: %w", err)
	}

	return &orm.MessageDocument{
		ID:          msg.SMessageID,
		ConvID:      msg.ConvID,
		Seq:         msg.Seq,
		SendID:      msg.SendID,
		RecvID:      msg.RecvID,
		ClientMsgID: msg.CMessaegID,
		Data:        data,
	}, nil
}

func messageFromDocument(ctx context.Context, doc *orm.MessageDocument) (*sdkws.MessageData, error) {
	msg := &sdkws.MessageData{}
	if err := msg.Unmarshal(doc.Data); err != nil {
		klog.CtxErrorf(ctx, "[MessageRepository] unmarshal message fact failed",
			"server_msg_id", doc.ID,
			"conv_id", doc.ConvID,
			"seq", doc.Seq,
			"error", err.Error())
		return nil, err
	}
	return msg, nil
}
