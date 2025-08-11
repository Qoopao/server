//
// Author : Ruanhuipeng
// Date   : 2025/05/26

package msggateway

import (
	"context"
	"encoding/json"

	// "github.com/cloudwego/kitex/client"
	// "github.com/cloudwego/kitex/transport"
	"github.com/openimsdk/tools/errs"
	"github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
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

type MessageHandler interface {
	SendMessage(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SdkWSResp, error)
	GetConvMsgList(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SdkWSResp, error)
	GetUserMsgList(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SdkWSResp, error)
}

type messageHandler struct {
	msgClient  messageservice.Client
	convClient conversationservice.Client
}

func NewMessageHandler(msgClient messageservice.Client, convClient conversationservice.Client) *messageHandler {
	return &messageHandler{
		msgClient:  msgClient,
		convClient: convClient,
	}
}

// SendMessage handles the sending of messages through gRPC. It unmarshals the request data,
// validates the message, and then sends it using the message RPC client.
func (g *messageHandler) SendMessage(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SdkWSResp, error) {
	var (
		sendMsgReq sdkws.SendMessageReq
		resp       *sdkws.SendMessageResp
		err        error
		respBody   []byte
	)

	if err = sendMsgReq.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "SendMessage: error unmarshaling message data", "action", "unmarshal", "dataType", "SendMessageReq")
	}

	if resp, err = g.msgClient.SendMessages(ctx, &sendMsgReq); err != nil {
		return nil, err
	}

	if respBody, err = resp.Marshal(nil); err != nil {
		return nil, errs.WrapMsg(err, "SendMessage: error marshaling message data", "action", "marshal", "dataType", "SendMessageResp")
	}

	return generateSdkWSResp(respBody, data), nil
}

func (g *messageHandler) GetConvMsgList(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SdkWSResp, error) {
	var (
		err            error
		request        *sdkws.FetchConvMessageListReq = &sdkws.FetchConvMessageListReq{}
		response       *sdkws.FetchConvMessageListResp
		responseBinary []byte
	)

	if err = request.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "GetConvMsgList: error unmarshaling message data", "action", "unmarshal", "dataType", "FetchConvMessageListReq")
	}

	if response, err = g.convClient.FetchConvMessageList(ctx, request); err != nil {
		return nil, errs.WrapMsg(err, "GetConvMsgList: error fetching conversation message list", "action", "fetch", "dataType", "FetchConvMessageListResp")
	}

	if responseBinary, err = response.Marshal(nil); err != nil {
		return nil, errs.WrapMsg(err, "GetConvMsgList: error marshaling message data", "action", "marshal", "dataType", "FetchConvMessageListResp")
	}

	return generateSdkWSResp(responseBinary, data), nil
}

func (g *messageHandler) GetUserMsgList(ctx context.Context, data *sdkws.SdkWSReq) (*sdkws.SdkWSResp, error) {
	var (
		err            error
		request        *sdkws.FetchUserMessageListReq = &sdkws.FetchUserMessageListReq{}
		response       *sdkws.FetchUserMessageListResp
		responseBinary []byte
	)

	if err = request.Unmarshal(data.Data); err != nil {
		return nil, errs.WrapMsg(err, "GetUserMsgList: error unmarshaling message data", "action", "unmarshal", "dataType", "FetchUserMessageListReq")
	}

	if response, err = g.convClient.FetchUserMessageList(ctx, request); err != nil {
		return nil, errs.WrapMsg(err, "GetUserMsgList: error fetching user message list", "action", "fetch", "dataType", "FetchUserMessageListResp")
	}

	if responseBinary, err = response.Marshal(nil); err != nil {
		return nil, errs.WrapMsg(err, "GetUserMsgList: error marshaling message data", "action", "marshal", "dataType", "FetchUserMessageListResp")
	}

	return generateSdkWSResp(responseBinary, data), nil
}

func generateSdkWSResp(data []byte, req *sdkws.SdkWSReq) *sdkws.SdkWSResp {
	return &sdkws.SdkWSResp{
		Data:      data,
		RequestId: req.RequestId,
		Token:     req.Token,
		UserID:    req.UserID,
		DeviceID:  req.DeviceID,
	}
}
