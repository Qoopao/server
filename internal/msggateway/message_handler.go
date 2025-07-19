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
	"github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
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
	SendMessage(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SendMessageResp, error)
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

// SendMessage handles the sending of messages through gRPC. It unmarshals the request data,
// validates the message, and then sends it using the message RPC client.
func (g *messageHandler) SendMessage(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SendMessageResp, error) {
	var sendMsgReq sdkws.SendMessageReq
	if err := sendMsgReq.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "SendMessage: error unmarshaling message data", "action", "unmarshal", "dataType", "SendMessageReq")
	}

	// if err := g.validate.Struct(&msgData); err != nil {
	// 	return nil, errs.WrapMsg(err, "SendMessage: message data validation failed", "action", "validate", "dataType", "MsgData")
	// }

	resp, err := g.msgClient.SendMessages(ctx, &sendMsgReq)
	if err != nil {
		println("SendMessage: error sending message", err.Error())
		return nil, err
	}
	return resp, nil
}
