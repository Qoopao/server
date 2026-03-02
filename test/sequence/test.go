package sequence

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	sequenceservice "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// SequenceClientTest 调用 sequence 服务，演示客户端请求链路是否正常
func SequenceClientTest() {
	// 1. 解析服务地址配置
	host := getenvOrDefault("SEQUENCE_SERVICE_HOST", "10.255.255.254")
	port := getenvOrDefault("SEQUENCE_SERVICE_PORT", "10300")
	destService := consts.SequenceServiceName

	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("[sequence-test] using addr=%s destService=%s", addr, destService)

	// 2. 创建 kitex 客户端
	cli, err := sequenceservice.NewClient(
		destService,
		client.WithHostPorts(addr),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: destService}),
	)
	if err != nil {
		log.Fatalf("[sequence-test] failed to create client: %v", err)
	}

	// 3. 构造请求并调用 GetNextSeq
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &sequencepb.GetNextSeqRequest{
		ConversationId: "test-conversation-1",
	}

	resp, err := cli.GetNextSeq(ctx, req)
	if err != nil {
		log.Fatalf("[sequence-test] GetNextSeq failed: %v", err)
	}

	// 4. 基础结果校验与输出
	if resp == nil {
		log.Fatalf("[sequence-test] GetNextSeq returned nil response")
	}

	if resp.ErrorCode != "" {
		log.Fatalf("[sequence-test] GetNextSeq business error: code=%s msg=%s", resp.ErrorCode, resp.ErrorMsg)
	}

	if resp.Seq <= 0 {
		log.Fatalf("[sequence-test] unexpected seq value: %d", resp.Seq)
	}

	log.Printf("[sequence-test] success: conversation_id=%s seq=%d", req.ConversationId, resp.Seq)
}

func getenvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
