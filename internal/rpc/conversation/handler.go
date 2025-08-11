package conversation

import (
	"context"

	sdkws "github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
)

// ConversationServiceImpl implements the last service interface defined in the IDL.
type ConversationServiceImpl struct {
	MessageDB controller.CommonMsgDatabase
}

// FetchConvMessageList implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListReq) (resp *sdkws.FetchConvMessageListResp, err error) {
	return s.fetchConvMessageList(ctx, req)
}

// FetchUserMessageList implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) FetchUserMessageList(ctx context.Context, req *sdkws.FetchUserMessageListReq) (resp *sdkws.FetchUserMessageListResp, err error) {
	return s.fetchUserMessageList(ctx, req)
}
