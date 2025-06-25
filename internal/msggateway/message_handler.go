//
// Author : Ruanhuipeng
// Date   : 2025/05/26

package msggateway

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/utils/jsonutil"
	"github.com/roc/roc-im-server/internal/kitex_gen/msg"
	"github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/push"
	"github.com/roc/roc-im-server/internal/kitex_gen/push/pushservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
)

const (
	TextPing = "ping"
	TextPong = "pong"
)

type TextMessage struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

type Req struct {
	ReqIdentifier int32  `json:"reqIdentifier" validate:"required"`
	Token         string `json:"token"`
	SendID        string `json:"sendID"        validate:"required"`
	OperationID   string `json:"operationID"   validate:"required"`
	MsgIncr       string `json:"msgIncr"       validate:"required"`
	Data          []byte `json:"data"`
}

func (r *Req) String() string {
	var tReq Req
	tReq.ReqIdentifier = r.ReqIdentifier
	tReq.Token = r.Token
	tReq.SendID = r.SendID
	tReq.OperationID = r.OperationID
	tReq.MsgIncr = r.MsgIncr
	return jsonutil.StructToJsonString(tReq)
}

var reqPool = sync.Pool{
	New: func() any {
		return new(Req)
	},
}

func getReq() *Req {
	req := reqPool.Get().(*Req)
	req.Data = nil
	req.MsgIncr = ""
	req.OperationID = ""
	req.ReqIdentifier = 0
	req.SendID = ""
	req.Token = ""
	return req
}

func freeReq(req *Req) {
	reqPool.Put(req)
}

type Resp struct {
	ReqIdentifier int32  `json:"reqIdentifier"`
	MsgIncr       string `json:"msgIncr"`
	OperationID   string `json:"operationID"`
	ErrCode       int    `json:"errCode"`
	ErrMsg        string `json:"errMsg"`
	Data          []byte `json:"data"`
}

func (r *Resp) String() string {
	var tResp Resp
	tResp.ReqIdentifier = r.ReqIdentifier
	tResp.MsgIncr = r.MsgIncr
	tResp.OperationID = r.OperationID
	tResp.ErrCode = r.ErrCode
	tResp.ErrMsg = r.ErrMsg
	return jsonutil.StructToJsonString(tResp)
}

type MessageHandler interface {
	GetSeq(ctx context.Context, data *Req) ([]byte, error)
	SendMessage(ctx context.Context, data *Req) ([]byte, error)
	SendSignalMessage(ctx context.Context, data *Req) ([]byte, error)
	PullMessageBySeqList(ctx context.Context, data *Req) ([]byte, error)
	GetConversationsHasReadAndMaxSeq(ctx context.Context, data *Req) ([]byte, error)
	GetSeqMessage(ctx context.Context, data *Req) ([]byte, error)
	UserLogout(ctx context.Context, data *Req) ([]byte, error)
	SetUserDeviceBackground(ctx context.Context, data *Req) ([]byte, bool, error)
	GetLastMessage(ctx context.Context, data *Req) ([]byte, error)
}

type messageHandler struct {
	validate   *validator.Validate
	msgClient  messageservice.Client
	pushClient pushservice.Client
}

func NewMessageHandler(validate *validator.Validate, msgClient messageservice.Client, pushClient pushservice.Client) *messageHandler {
	return &messageHandler{
		validate:   validate,
		msgClient:  msgClient,
		pushClient: pushClient,
	}
}

func (g *messageHandler) GetSeq(ctx context.Context, data *Req) ([]byte, error) {
	req := sdkws.GetMaxSeqReq{}
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "GetSeq: error unmarshaling request", "action", "unmarshal", "dataType", "GetMaxSeqReq")
	}
	if err := g.validate.Struct(&req); err != nil {
		return nil, errs.WrapMsg(err, "GetSeq: validation failed", "action", "validate", "dataType", "GetMaxSeqReq")
	}
	resp, err := g.msgClient.GetMaxSeq(ctx, &req)
	if err != nil {
		return nil, err
	}
	// c, err := resp.Marshal(nil)
	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "GetSeq: error marshaling response", "action", "marshal", "dataType", "GetMaxSeqResp")
	}
	return c, nil
}

// SendMessage handles the sending of messages through gRPC. It unmarshals the request data,
// validates the message, and then sends it using the message RPC client.
func (g *messageHandler) SendMessage(ctx context.Context, data *Req) ([]byte, error) {
	var msgData sdkws.MsgData
	if err := msgData.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "SendMessage: error unmarshaling message data", "action", "unmarshal", "dataType", "MsgData")
	}

	// if err := g.validate.Struct(&msgData); err != nil {
	// 	return nil, errs.WrapMsg(err, "SendMessage: message data validation failed", "action", "validate", "dataType", "MsgData")
	// }

	req := msg.SendMsgReq{MsgData: &msgData}
	resp, err := g.msgClient.SendMsg(ctx, &req)
	if err != nil {
		return nil, err
	}

	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "SendMessage: error marshaling response", "action", "marshal", "dataType", "SendMsgResp")
	}

	return c, nil
}

func (g *messageHandler) SendSignalMessage(ctx context.Context, data *Req) ([]byte, error) {
	resp, err := g.msgClient.SendMsg(ctx, nil)
	if err != nil {
		return nil, err
	}
	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "error marshaling response", "action", "marshal", "dataType", "SendMsgResp")
	}
	return c, nil
}

func (g *messageHandler) PullMessageBySeqList(ctx context.Context, data *Req) ([]byte, error) {
	req := sdkws.PullMessageBySeqsReq{}
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "err proto unmarshal", "action", "unmarshal", "dataType", "PullMessageBySeqsReq")
	}
	if err := g.validate.Struct(data); err != nil {
		return nil, errs.WrapMsg(err, "validation failed", "action", "validate", "dataType", "PullMessageBySeqsReq")
	}
	resp, err := g.msgClient.PullMessageBySeqs(ctx, &req)
	if err != nil {
		return nil, err
	}
	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "error marshaling response", "action", "marshal", "dataType", "PullMessageBySeqsResp")
	}
	return c, nil
}

func (g *messageHandler) GetConversationsHasReadAndMaxSeq(ctx context.Context, data *Req) ([]byte, error) {
	req := msg.GetConversationsHasReadAndMaxSeqReq{}
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "err proto unmarshal", "action", "unmarshal", "dataType", "GetConversationsHasReadAndMaxSeq")
	}
	if err := g.validate.Struct(data); err != nil {
		return nil, errs.WrapMsg(err, "validation failed", "action", "validate", "dataType", "GetConversationsHasReadAndMaxSeq")
	}
	resp, err := g.msgClient.GetConversationsHasReadAndMaxSeq(ctx, &req)
	if err != nil {
		return nil, err
	}
	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "error marshaling response", "action", "marshal", "dataType", "GetConversationsHasReadAndMaxSeq")
	}
	return c, nil
}

func (g *messageHandler) GetSeqMessage(ctx context.Context, data *Req) ([]byte, error) {
	req := msg.GetSeqMessageReq{}
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "error unmarshaling request", "action", "unmarshal", "dataType", "GetSeqMessage")
	}
	if err := g.validate.Struct(data); err != nil {
		return nil, errs.WrapMsg(err, "validation failed", "action", "validate", "dataType", "GetSeqMessage")
	}
	resp, err := g.msgClient.GetSeqMessage(ctx, &req)
	if err != nil {
		return nil, err
	}
	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "error marshaling response", "action", "marshal", "dataType", "GetSeqMessage")
	}
	return c, nil
}

func (g *messageHandler) UserLogout(ctx context.Context, data *Req) ([]byte, error) {
	req := push.DelUserPushTokenReq{}
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "error unmarshaling request", "action", "unmarshal", "dataType", "DelUserPushTokenReq")
	}
	resp, err := g.pushClient.DelUserPushToken(ctx, &req)
	if err != nil {
		return nil, err
	}
	c, err := resp.Marshal(nil)
	if err != nil {
		return nil, errs.WrapMsg(err, "error marshaling response", "action", "marshal", "dataType", "DelUserPushTokenResp")
	}
	return c, nil
}

func (g *messageHandler) SetUserDeviceBackground(ctx context.Context, data *Req) ([]byte, bool, error) {
	req := sdkws.SetAppBackgroundStatusReq{}
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, false, errs.WrapMsg(err, "error unmarshaling request", "action", "unmarshal", "dataType", "SetAppBackgroundStatusReq")
	}
	if err := g.validate.Struct(data); err != nil {
		return nil, false, errs.WrapMsg(err, "validation failed", "action", "validate", "dataType", "SetAppBackgroundStatusReq")
	}
	return nil, req.IsBackground, nil
}

func (g *messageHandler) GetLastMessage(ctx context.Context, data *Req) ([]byte, error) {
	var req msg.GetLastMessageReq
	if err := req.Unmarshal(data.Data); err != nil {
		return nil, err
	}
	resp, err := g.msgClient.GetLastMessage(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp.Marshal(nil)
}
