package push

import (
	"context"
	push "github.com/roc/roc-im-server/internal/kitex_gen/push"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct{}

// PushMsg implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushMsg(ctx context.Context, req *push.PushMsgReq) (resp *push.PushMsgResp, err error) {
	// TODO: Your code here...
	return
}

// DelUserPushToken implements the PushServiceImpl interface.
func (s *PushServiceImpl) DelUserPushToken(ctx context.Context, req *push.DelUserPushTokenReq) (resp *push.DelUserPushTokenResp, err error) {
	// TODO: Your code here...
	return
}
