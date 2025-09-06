//
// Author : Ruanhuipeng
// Date   : 2025/05/26

package msggateway

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/transport"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/stringutil"
	"github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"go.opentelemetry.io/otel"
)

var (
	ErrConnClosed                = errs.New("conn has closed")
	ErrNotSupportMessageProtocol = errs.New("not support message protocol")
	ErrClientClosed              = errs.New("client actively close the connection")
	ErrPanic                     = errs.New("panic error")
)

type Client struct {
	w            *sync.Mutex
	conn         LongConn
	PlatformID   int    `json:"platformID"`
	IsCompress   bool   `json:"isCompress"`
	UserID       string `json:"userID"`
	IsBackground bool   `json:"isBackground"`
	SDKType      string `json:"sdkType"`
	Encoder      Encoder
	ctx          *UserConnContext
	closed       atomic.Bool
	closedErr    error
	token        string
	hbCtx        context.Context
	hbCancel     context.CancelFunc
	subLock      *sync.Mutex
	subUserIDs   map[string]struct{} // client conn subscription list

	messsageHandler MessageHandler
	composer        Compressor

	closeCallback func(err error)
}

// ResetClient updates the client's state with new connection and context information.
func (c *Client) ResetClient(ctx *UserConnContext, conn LongConn) {
	c.w = new(sync.Mutex)
	c.conn = conn
	c.PlatformID = stringutil.StringToInt(ctx.GetPlatformID())
	c.IsCompress = ctx.GetCompression()
	c.IsBackground = ctx.GetBackground()
	c.UserID = ctx.GetUserID()
	c.ctx = ctx
	c.IsBackground = false
	c.closed.Store(false)
	c.closedErr = nil
	c.token = ctx.GetToken()
	c.SDKType = ctx.GetSDKType()
	c.hbCtx, c.hbCancel = context.WithCancel(c.ctx)
	c.subLock = new(sync.Mutex)
	if c.subUserIDs != nil {
		clear(c.subUserIDs)
	}
	if c.SDKType == GoSDK {
		c.Encoder = NewGobEncoder()
	} else {
		c.Encoder = NewJsonEncoder()
	}
	c.subUserIDs = make(map[string]struct{})

	// 使用服务发现获取 msg 服务客户端
	msgRpc, err := GetMsgServiceClient()
	if err != nil {
		// 降级到硬编码地址
		msgRpc, _ = messageservice.NewClient("example_service", client.WithHostPorts("0.0.0.0:10100"), client.WithTransportProtocol(transport.GRPC))
	}

	// 使用服务发现获取 conversation 服务客户端
	convRpc, err := GetConversationServiceClient()
	if err != nil {
		// 降级到硬编码地址
		convRpc, _ = conversationservice.NewClient("example_service", client.WithHostPorts("0.0.0.0:10200"), client.WithTransportProtocol(transport.GRPC))
	}

	c.messsageHandler = NewMessageHandler(msgRpc, convRpc)
}

func (c *Client) pingHandler(appData string) error {
	if err := c.conn.SetReadDeadline(pongWait); err != nil {
		return err
	}

	log.ZDebug(c.ctx, "ping Handler Success.", "appData", appData)
	return c.writePongMsg(appData)
}

func (c *Client) pongHandler(_ string) error {
	if err := c.conn.SetReadDeadline(pongWait); err != nil {
		return err
	}
	return nil
}

// readMessage continuously reads messages from the connection.
func (c *Client) readMessage() {
	defer func() {
		if r := recover(); r != nil {
			c.closedErr = ErrPanic
			// log.ZPanic(c.ctx, "socket have panic err:", errs.ErrPanic(r))
		}
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(pongWait)
	c.conn.SetPongHandler(c.pongHandler)
	c.conn.SetPingHandler(c.pingHandler)
	c.activeHeartbeat(c.hbCtx)

	for {
		log.ZDebug(c.ctx, "readMessage")
		messageType, message, returnErr := c.conn.ReadMessage()
		if returnErr != nil {
			log.ZWarn(c.ctx, "readMessage", returnErr, "messageType", messageType)
			c.closedErr = returnErr
			c.closeCallback(returnErr)
			return
		}

		log.ZDebug(c.ctx, "readMessage", "messageType", messageType)
		if c.closed.Load() {
			// The scenario where the connection has just been closed, but the coroutine has not exited
			c.closedErr = ErrConnClosed
			c.closeCallback(c.closedErr)
			return
		}

		switch messageType {
		case MessageBinary:
			_ = c.conn.SetReadDeadline(pongWait)
			parseDataErr := c.handleMessage(message)
			if parseDataErr != nil {
				c.closedErr = parseDataErr
				return
			}
		case MessageText:
			_ = c.conn.SetReadDeadline(pongWait)
			parseDataErr := c.handlerTextMessage(message)
			if parseDataErr != nil {
				c.closedErr = parseDataErr
				return
			}
		case MessagePing:
			err := c.writePongMsg("")
			log.ZError(c.ctx, "writePongMsg", err)

		case MessageClose:
			c.closedErr = ErrClientClosed
			return

		default:
		}
	}
}

// handleMessage processes a single message received by the client.
func (c *Client) handleMessage(message []byte) error {

	var (
		err      error
		data     []byte
		resp     *sdkws.SdkWSResp
		sdkwsReq *sdkws.SdkWSReq = &sdkws.SdkWSReq{}
		ctx      context.Context = context.Background()
	)

	if err = sdkwsReq.Unmarshal(message); err != nil {
		log.ZError(ctx, "handleMessage", err)
		return err
	}

	// 在这里就开始追踪
	ctx, span := otel.Tracer("msggateway").Start(ctx, "handleMessage")
	defer span.End()

	klog.CtxDebugf(ctx, "call handleMessage", "type", sdkwsReq.Type)

	switch sdkwsReq.Type {

	case WSSendMessage:
		resp, err = c.messsageHandler.SendMessage(ctx, sdkwsReq)

	case WSPullConvMsgList:
		resp, err = c.messsageHandler.GetConvMsgList(ctx, sdkwsReq)

	case WSPullUserMsgList:
		resp, err = c.messsageHandler.GetUserMsgList(ctx, sdkwsReq)

	}

	if err != nil {
		log.ZError(ctx, "handleMessage", err)
		return err
	}

	// 序列化
	if data, err = resp.Marshal(nil); err != nil {
		log.ZError(ctx, "handleMessage", err)
		return err
	}

	// 回复
	if err = c.replyMessage(ctx, data); err != nil {
		log.ZError(ctx, "replyMessage", err)
	}

	return nil
}

func (c *Client) close() {
	c.w.Lock()
	defer c.w.Unlock()
	if c.closed.Load() {
		return
	}
	c.closed.Store(true)
	c.conn.Close()
	c.hbCancel() // Close server-initiated heartbeat.

	// todo
	// c.longConnServer.UnRegister(c)
}

func (c *Client) replyMessage(ctx context.Context, resp []byte) error {
	return c.conn.WriteMessage(MessageBinary, resp)
}

func (c *Client) PushMessage(ctx context.Context, msgData *sdkws.MsgData) error {
	msgData.DStatus = 1 // 实时消息
	msg := sdkws.PushMessages{
		Msgs: []*sdkws.MessageUnion{{Msg: msgData, IsCmd: false}},
	}

	log.ZDebug(ctx, "PushMessage", "msg", &msg)
	data, err := msgData.Marshal(nil)
	if err != nil {
		return err
	}

	resp := sdkws.SdkWSResp{
		Data: data,
		Type: 104,
	}

	respBinary, _ := resp.Marshal(nil)
	return c.conn.WriteMessage(MessageBinary, respBinary)
}

// Actively initiate Heartbeat when platform in Web.
func (c *Client) activeHeartbeat(ctx context.Context) {
	// if c.PlatformID == constant.LinuxPlatformID {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.ZPanic(ctx, "activeHeartbeat Panic", errs.ErrPanic(r))
			}
		}()
		log.ZDebug(ctx, "server initiative send heartbeat start.")
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.writePingMsg(); err != nil {
					log.ZWarn(c.ctx, "send Ping Message error.", err)
					return
				}
			case <-c.hbCtx.Done():
				return
			}
		}
	}()
	// }
}

func (c *Client) writePingMsg() error {
	if c.closed.Load() {
		return nil
	}

	c.w.Lock()
	defer c.w.Unlock()

	err := c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(MessagePing, nil)
}

func (c *Client) writePongMsg(appData string) error {
	log.ZDebug(c.ctx, "write Pong Msg in Server", "appData", appData)
	if c.closed.Load() {
		log.ZWarn(c.ctx, "is closed in server", nil, "appdata", appData, "closed err", c.closedErr)
		return nil
	}

	c.w.Lock()
	defer c.w.Unlock()

	err := c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		log.ZWarn(c.ctx, "SetWriteDeadline in Server have error", errs.Wrap(err), "writeWait", writeWait, "appData", appData)
		return errs.Wrap(err)
	}
	err = c.conn.WriteMessage(MessagePong, []byte(appData))
	if err != nil {
		log.ZWarn(c.ctx, "Write Message have error", errs.Wrap(err), "Pong msg", MessagePong)
	}

	return errs.Wrap(err)
}

func (c *Client) handlerTextMessage(b []byte) error {
	var msg TextMessage
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case TextPong:
		return nil
	case TextPing:
		msg.Type = TextPong
		msgData, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		c.w.Lock()
		defer c.w.Unlock()
		if err := c.conn.SetWriteDeadline(writeWait); err != nil {
			return err
		}
		return c.conn.WriteMessage(MessageText, msgData)
	default:
		return fmt.Errorf("not support message type %s", msg.Type)
	}
}
