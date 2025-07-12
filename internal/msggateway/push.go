package msggateway

import (
	"context"
	"net"

	"github.com/cloudwego/kitex/server"
	push "github.com/roc/roc-im-server/internal/kitex_gen/push"
	pushservice "github.com/roc/roc-im-server/internal/kitex_gen/push/pushservice"
)

func startPushService(wsServer *WsServer) {

	svr := pushservice.NewServer(&PushServiceImpl{
		wsServer: wsServer,
	}, server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 10300}))

	err := svr.Run()

	if err != nil {
		panic(err.Error())
	}
}

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
	wsServer *WsServer
}

// PushMsg implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushMsg(ctx context.Context, req *push.PushMsgReq) (resp *push.PushMsgResp, err error) {
	if len(req.UserIDs) == 0 {
		return &push.PushMsgResp{}, nil
	}
	err = s.wsServer.pushToUser(ctx, req.UserIDs[0], req.MsgData)
	return &push.PushMsgResp{}, err
}

// DelUserPushToken implements the PushServiceImpl interface.
func (s *PushServiceImpl) DelUserPushToken(ctx context.Context, req *push.DelUserPushTokenReq) (resp *push.DelUserPushTokenResp, err error) {
	// TODO: Your code here...
	return
}
