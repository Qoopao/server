package kafaka_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/roc/roc-im-server/tools/mq" // 替换为实际的包路径
)

func KfakTest() {
	// 1. 创建 MQ 实例
	mqInstance, err := mq.NewSaramaMQ([]string{"localhost:9092"})
	if err != nil {
		log.Fatalf("Failed to create MQ: %v", err)
	}
	defer mqInstance.Close()

	// 2. 创建主题
	topic := "orders"
	dlqTopic := "orders-dlq"

	if err := mqInstance.CreateTopic(context.Background(), topic, 3); err != nil {
		log.Printf("Failed to create topic '%s': %v", topic, err)
	}

	if err := mqInstance.CreateTopic(context.Background(), dlqTopic, 1); err != nil {
		log.Printf("Failed to create DLQ topic '%s': %v", dlqTopic, err)
	}

	// 3. 启动消费者
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// 启动正常订单处理器
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startOrderProcessor(ctx, mqInstance, topic, dlqTopic); err != nil {
			log.Printf("Order processor failed: %v", err)
		}
	}()

	// 启动死信队列处理器
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startDLQProcessor(ctx, mqInstance, dlqTopic); err != nil {
			log.Printf("DLQ processor failed: %v", err)
		}
	}()

	// 4. 启动生产者
	wg.Add(1)
	go func() {
		defer wg.Done()
		produceOrders(ctx, mqInstance, topic)
	}()

	// 5. 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Application started. Press Ctrl+C to exit.")

	select {
	case sig := <-sigCh:
		log.Printf("Received signal: %s. Shutting down...", sig)
		cancel() // 通知所有协程停止
	case <-ctx.Done():
	}

	// 等待所有协程完成
	wg.Wait()
	log.Println("Application shutdown complete.")
}

// startOrderProcessor 启动订单处理器
func startOrderProcessor(ctx context.Context, mqi mq.MQ, topic, dlqTopic string) error {
	log.Printf("Starting order processor for topic '%s'...", topic)

	// 订阅主题并处理消息
	err := mqi.Subscribe(topic, func(ctx context.Context, msg *mq.Message) error {
		log.Printf("Processing order: %s", string(msg.Body))

		// 模拟处理逻辑 - 随机失败
		if time.Now().UnixNano()%5 == 0 { // 20% 失败率
			log.Printf("Order processing failed: %s", string(msg.Body))
			return fmt.Errorf("processing failed for order: %s", string(msg.Body))
		}

		// 模拟处理时间
		time.Sleep(100 * time.Millisecond)
		log.Printf("Order processed successfully: %s", string(msg.Body))
		return nil
	},
		mq.WithGroupID("order-processor"),
		mq.WithConcurrency(3),
		mq.WithRetryPolicy(3, 500*time.Millisecond),
		mq.WithMaxBackoff(3*time.Second),
		mq.WithDeadLetterQueue(dlqTopic, 3),
		mq.WithMessageTimeout(2*time.Second),
	)

	if err != nil {
		return fmt.Errorf("failed to subscribe to topic '%s': %w", topic, err)
	}

	// 等待上下文取消
	<-ctx.Done()

	// 取消订阅
	if err := mqi.Unsubscribe(topic); err != nil {
		log.Printf("Failed to unsubscribe from topic '%s': %v", topic, err)
	}

	return nil
}

// startDLQProcessor 启动死信队列处理器
func startDLQProcessor(ctx context.Context, mq1 mq.MQ, dlqTopic string) error {
	log.Printf("Starting DLQ processor for topic '%s'...", dlqTopic)

	// 订阅死信队列
	err := mq1.Subscribe(dlqTopic, func(ctx context.Context, msg *mq.Message) error {
		log.Printf("Processing DLQ message: %s", string(msg.Body))

		// 这里可以添加处理死信消息的逻辑
		// 例如：记录日志、发送警报、人工干预等
		log.Printf("DLQ message received and processed: %s", string(msg.Body))
		return nil
	},
		mq.WithGroupID("dlq-processor"),
		mq.WithAutoAck(true),
	)

	if err != nil {
		return fmt.Errorf("failed to subscribe to DLQ topic '%s': %w", dlqTopic, err)
	}

	// 等待上下文取消
	<-ctx.Done()

	// 取消订阅
	if err := mq1.Unsubscribe(dlqTopic); err != nil {
		log.Printf("Failed to unsubscribe from DLQ topic '%s': %v", dlqTopic, err)
	}

	return nil
}

// produceOrders 生产订单消息
func produceOrders(ctx context.Context, mq1 mq.MQ, topic string) {
	orderCounter := 1
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("Starting order producer for topic '%s'...", topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Order producer stopped")
			return
		case <-ticker.C:
			order := fmt.Sprintf("Order-%d", orderCounter)
			msg := &mq.Message{
				Topic: topic,
				Body:  []byte(order),
			}

			if err := mq1.Publish(ctx, msg); err != nil {
				log.Printf("Failed to publish order '%s': %v", order, err)
			} else {
				log.Printf("Published order: %s", order)
			}

			orderCounter++

			// 每10个订单批量发布一次
			if orderCounter%10 == 0 {
				batch := make([]*mq.Message, 0, 5)
				for i := 0; i < 5; i++ {
					batchOrder := fmt.Sprintf("Batch-Order-%d", orderCounter+i)
					batch = append(batch, &mq.Message{
						Topic: topic,
						Body:  []byte(batchOrder),
					})
				}

				if err := mq1.BatchPublish(ctx, batch); err != nil {
					log.Printf("Failed to publish batch orders: %v", err)
				} else {
					log.Printf("Published batch of %d orders", len(batch))
				}
				orderCounter += 5
			}
		}
	}
}
