package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/rhp-QE/roc-foundation-service/long_connection_service/src/start"
	convMsgConsumer "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer"
	backserviceRpc "github.com/rhp-QE/roc-im-server/src/rpc/backservice"
	conversationRpc "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service"
	messageRpc "github.com/rhp-QE/roc-im-server/src/rpc/message_service"
	sequenceRpc "github.com/rhp-QE/roc-im-server/src/rpc/sequence_service"
)

// Service 定义服务接口
type Service struct {
	Name  string
	Start func() error
}

// ServiceWithoutError 定义不返回错误的服务接口
type ServiceWithoutError struct {
	Name  string
	Start func()
}

// startLongConnectionService 启动长连接服务的包装函数
func startLongConnectionService() error {
	if configPath := os.Getenv("LONG_CONNECTION_CONFIG"); configPath != "" {
		return start.Start(configPath)
	}

	_, file, _, ok := runtime.Caller(0)
	if ok {
		return start.Start(filepath.Join(filepath.Dir(file), "config.yaml"))
	}

	return start.Start("cmd/all-services/config.yaml")
}

func main() {
	log.Println("启动所有 IM 服务...")

	// 定义长连接服务（优先启动）
	longConnectionService := Service{
		Name:  "Long Connection Service",
		Start: startLongConnectionService,
	}

	// 定义其他需要启动的服务
	services := []Service{
		{
			Name:  "Message Service",
			Start: messageRpc.Start,
		},
		{
			Name:  "Conversation Service",
			Start: conversationRpc.Start,
		},
		{
			Name:  "Sequence Service",
			Start: sequenceRpc.Start,
		},
		{
			Name:  "Conversation Message Consumer",
			Start: convMsgConsumer.Start,
		},
	}

	servicesWithoutError := []ServiceWithoutError{
		{
			Name:  "Back Service",
			Start: backserviceRpc.Start,
		},
	}

	// 创建错误通道
	errChan := make(chan error, len(services)+1)

	// 优先启动长连接服务（直接依赖 Start 的返回值）
	log.Printf("正在启动 %s...", longConnectionService.Name)
	go func() {
		if err := longConnectionService.Start(); err != nil {
			errChan <- fmt.Errorf("%s 启动失败: %v", longConnectionService.Name, err)
		}
	}()

	time.Sleep(3 * time.Second)

	// 启动返回 error 的其他服务：不再用 sleep + 乐观判断，而是直接看 Start 的返回值
	for _, svc := range services {
		go func(service Service) {
			log.Printf("正在启动 %s...", service.Name)
			if err := service.Start(); err != nil {
				errChan <- fmt.Errorf("%s 启动失败: %v", service.Name, err)
			}
		}(svc)
	}

	// 启动不返回 error 的服务（保持原有行为）
	for _, svc := range servicesWithoutError {
		go func(service ServiceWithoutError) {
			log.Printf("正在启动 %s...", service.Name)
			service.Start()
		}(svc)
	}

	// 设置优雅关闭
	setupGracefulShutdown()

	// 任何一个 Start 返回错误，都认为整个进程需要退出
	go func() {
		if err := <-errChan; err != nil {
			log.Fatalf("服务启动或运行失败: %v", err)
		}
	}()

	log.Println("所有服务启动任务已提交，等待关闭信号...")

	// 等待关闭信号
	waitForShutdown()
}

// setupGracefulShutdown 设置优雅关闭处理
func setupGracefulShutdown() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Printf("接收到关闭信号: %v，正在停止所有服务...", sig)

		// 再次接收信号时强制退出
		go func() {
			sig := <-c
			log.Printf("再次接收到关闭信号: %v，强制退出", sig)
			os.Exit(1)
		}()

		// 这里可以添加服务清理逻辑
		// 例如：关闭数据库连接、取消服务注册等
		log.Println("服务清理完成")

		log.Println("所有服务已停止")
		os.Exit(0)
	}()
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("接收到关闭信号: %v，所有服务正在关闭...", sig)

	// 阻塞等待优雅关闭处理完成
	select {}
}
