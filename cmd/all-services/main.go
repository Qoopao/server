package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
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
	return start.Start("")
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

	// 创建错误通道和启动状态跟踪
	errChan := make(chan error, len(services)+1)
	startedServices := make(chan string, len(services)+len(servicesWithoutError)+1)
	doneChan := make(chan bool, len(servicesWithoutError))

	// 优先启动长连接服务
	log.Printf("正在启动 %s...", longConnectionService.Name)
	longConnStartTime := time.Now()
	longConnErrChan := make(chan error, 1)
	go func() {
		longConnErrChan <- longConnectionService.Start()
	}()

	// 等待长连接服务启动（给服务一些时间完成初始化和注册）
	time.Sleep(3 * time.Second)

	// 检查长连接服务是否启动失败
	select {
	case err := <-longConnErrChan:
		if err != nil {
			log.Fatalf("%s 启动失败: %v", longConnectionService.Name, err)
		}
	default:
		// 服务还在运行，认为启动成功
	}

	longConnDuration := time.Since(longConnStartTime)
	log.Printf("%s 启动成功 (耗时: %v)", longConnectionService.Name, longConnDuration)
	startedServices <- longConnectionService.Name

	// 等待2秒后再启动其他服务
	log.Println("等待 2 秒后启动其他服务...")
	time.Sleep(2 * time.Second)

	// 启动返回 error 的其他服务
	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(service Service) {
			defer wg.Done()
			log.Printf("正在启动 %s...", service.Name)
			startTime := time.Now()

			// 在 goroutine 中启动服务（因为 Start() 是阻塞的）
			serviceErrChan := make(chan error, 1)
			go func() {
				serviceErrChan <- service.Start()
			}()

			// 等待服务启动（给服务一些时间完成初始化和注册）
			time.Sleep(3 * time.Second)

			// 检查服务是否启动失败（非阻塞检查）
			select {
			case err := <-serviceErrChan:
				if err != nil {
					errChan <- fmt.Errorf("%s 启动失败: %v", service.Name, err)
					return
				}
			default:
				// 服务还在运行，认为启动成功
			}

			duration := time.Since(startTime)
			log.Printf("%s 启动成功 (耗时: %v)", service.Name, duration)
			startedServices <- service.Name
		}(svc)
	}

	// 启动不返回 error 的服务
	for _, svc := range servicesWithoutError {
		go func(service ServiceWithoutError) {
			log.Printf("正在启动 %s...", service.Name)
			startTime := time.Now()
			// 在 goroutine 中启动服务（因为 Start() 可能是阻塞的）
			go func() {
				service.Start()
			}()
			// 等待一小段时间确保服务已启动，然后发送启动信号
			time.Sleep(2 * time.Second)
			duration := time.Since(startTime)
			log.Printf("%s 启动成功 (耗时: %v)", service.Name, duration)
			startedServices <- service.Name
			doneChan <- true
		}(svc)
	}

	// 等待所有服务启动完成或出错
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 设置优雅关闭
	setupGracefulShutdown()

	// 监听启动状态和错误
	totalServices := len(services) + len(servicesWithoutError) + 1 // +1 是长连接服务
	startedCount := 0

	for {
		select {
		case err := <-errChan:
			log.Fatalf("服务启动失败: %v", err)
		case serviceName := <-startedServices:
			startedCount++
			log.Printf("服务启动进度: %d/%d (%s)", startedCount, totalServices, serviceName)
			if startedCount == totalServices {
				log.Println("所有服务启动完成！")
				goto waitForShutdown
			}
		case <-time.After(120 * time.Second):
			log.Fatalf("服务启动超时（120秒）")
		}
	}

waitForShutdown:
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
