package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/roc/roc-im-server/internal/kitex_gen/msg"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/internal/msggateway"
)

// MultiService_main 多服务调用示例主函数
func MultiService_main() {
	fmt.Println("=== 多服务调用示例 ===")

	// 示例1: 基本服务调用
	fmt.Println("\n1. 基本服务调用")
	basicServiceCall()

	// 示例2: 一致性哈希调用
	fmt.Println("\n2. 一致性哈希调用")
	consistentHashCall()

	// 示例3: 服务刷新
	fmt.Println("\n3. 服务刷新")
	serviceRefresh()

	fmt.Println("\n示例完成")
}

// basicServiceCall 基本服务调用示例
func basicServiceCall() {
	ctx := context.Background()

	// 调用消息服务
	msgClient, err := msggateway.GetMsgServiceClient()
	if err != nil {
		log.Printf("获取消息服务客户端失败: %v", err)
		return
	}

	// 消息服务调用示例
	msgReq := &sdkws.GetMaxSeqReq{
		UserID: "user123",
	}
	msgResp, err := msgClient.GetMaxSeq(ctx, msgReq)
	if err != nil {
		log.Printf("消息服务调用失败: %v", err)
	} else {
		fmt.Printf("消息服务调用成功: MaxSeqs = %v\n", msgResp.GetMaxSeqs())
	}

	// 会话服务调用示例 - 使用消息服务中的会话相关方法
	convReq := &msg.GetConversationMaxSeqReq{
		ConversationID: "conv123",
	}
	convResp, err := msgClient.GetConversationMaxSeq(ctx, convReq)
	if err != nil {
		log.Printf("会话服务调用失败: %v", err)
	} else {
		fmt.Printf("会话服务调用成功: MaxSeq = %d\n", convResp.GetMaxSeq())
	}
}

// consistentHashCall 一致性哈希调用示例
func consistentHashCall() {
	ctx := context.Background()

	// 模拟不同用户的请求
	userIDs := []string{"user123", "user456", "user789"}

	for _, userID := range userIDs {
		fmt.Printf("用户 %s 的请求:\n", userID)

		// 使用一致性哈希获取消息服务客户端
		msgClient, err := msggateway.GetMsgServiceClientWithKey(userID)
		if err != nil {
			log.Printf("获取消息服务客户端失败: %v", err)
			continue
		}

		// 调用消息服务
		msgReq := &sdkws.GetMaxSeqReq{
			UserID: userID,
		}
		_, err = msgClient.GetMaxSeq(ctx, msgReq)
		if err != nil {
			log.Printf("消息服务调用失败: %v", err)
		} else {
			fmt.Printf("  消息服务调用成功\n")
		}

		// 调用会话相关方法（通过消息服务）
		convReq := &msg.GetConversationMaxSeqReq{
			ConversationID: fmt.Sprintf("conv_%s", userID),
		}
		_, err = msgClient.GetConversationMaxSeq(ctx, convReq)
		if err != nil {
			log.Printf("会话服务调用失败: %v", err)
		} else {
			fmt.Printf("  会话服务调用成功\n")
		}
	}
}

// serviceRefresh 服务刷新示例
func serviceRefresh() {
	fmt.Println("刷新消息服务实例缓存...")
	err := msggateway.RefreshMsgService()
	if err != nil {
		log.Printf("刷新消息服务失败: %v", err)
	} else {
		fmt.Println("消息服务刷新成功")
	}

	fmt.Println("刷新会话服务实例缓存...")
	err = msggateway.RefreshConversationService()
	if err != nil {
		log.Printf("刷新会话服务失败: %v", err)
	} else {
		fmt.Println("会话服务刷新成功")
	}

	// 获取服务发现客户端查看所有服务
	client := msggateway.GetServiceDiscoveryClient()
	if client != nil {
		services := client.GetAllServices()
		fmt.Printf("当前注册的服务:\n")
		for serviceName, instances := range services {
			fmt.Printf("  %s: %d 个实例\n", serviceName, len(instances))
			for _, instance := range instances {
				fmt.Printf("    - %s:%d (状态: %v, 权重: %d)\n",
					instance.Host, instance.Port, instance.Status, instance.Weight)
			}
		}
	}
}

// 模拟服务调用的辅助函数
func simulateServiceCall(serviceName string, duration time.Duration) {
	fmt.Printf("模拟 %s 服务调用，耗时 %v...\n", serviceName, duration)
	time.Sleep(duration)
	fmt.Printf("%s 服务调用完成\n", serviceName)
}
