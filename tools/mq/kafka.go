package mq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// SaramaMQ 使用 sarama 实现 MQ 接口
type SaramaMQ struct {
	brokers    []string
	config     *sarama.Config
	admin      sarama.ClusterAdmin
	producers  map[string]sarama.SyncProducer
	consumers  map[string]*consumerGroupHandler
	producerMu sync.Mutex
	consumerMu sync.Mutex
	wg         sync.WaitGroup
}

type consumerGroupHandler struct {
	mq          *SaramaMQ
	topic       string
	handler     HandlerFunc
	options     SubscribeOptions
	group       sarama.ConsumerGroup
	ctx         context.Context
	cancel      context.CancelFunc
	dlqProducer sarama.SyncProducer
}

// NewSaramaMQ 创建新的 SaramaMQ 实例
func NewSaramaMQ(brokers []string) (*SaramaMQ, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0 // 使用兼容性较好的版本
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 创建集群管理客户端
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	return &SaramaMQ{
		brokers:   brokers,
		config:    config,
		admin:     admin,
		producers: make(map[string]sarama.SyncProducer),
		consumers: make(map[string]*consumerGroupHandler),
	}, nil
}

func (s *SaramaMQ) Publish(ctx context.Context, msg *Message) error {
	producer, err := s.getProducer(msg.Topic)
	if err != nil {
		return err
	}

	kmsg := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(msg.Body),
		Key:   sarama.StringEncoder(msg.Key),
	}

	_, _, err = producer.SendMessage(kmsg)
	return err
}

func (s *SaramaMQ) BatchPublish(ctx context.Context, msgs []*Message) error {
	if len(msgs) == 0 {
		return nil
	}

	// 按主题分组消息
	topicMsgs := make(map[string][]*sarama.ProducerMessage)
	for _, msg := range msgs {
		kmsg := &sarama.ProducerMessage{
			Topic: msg.Topic,
			Value: sarama.ByteEncoder(msg.Body),
		}
		topicMsgs[msg.Topic] = append(topicMsgs[msg.Topic], kmsg)
	}

	// 对每个主题批量发布
	for topic, msgs := range topicMsgs {
		producer, err := s.getProducer(topic)
		if err != nil {
			return err
		}

		if err := producer.SendMessages(msgs); err != nil {
			return fmt.Errorf("failed to send messages to topic %s: %w", topic, err)
		}
	}

	return nil
}

func (s *SaramaMQ) getProducer(topic string) (sarama.SyncProducer, error) {
	s.producerMu.Lock()
	defer s.producerMu.Unlock()

	if producer, exists := s.producers[topic]; exists {
		return producer, nil
	}

	producer, err := sarama.NewSyncProducer(s.brokers, s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer for topic %s: %w", topic, err)
	}

	s.producers[topic] = producer
	return producer, nil
}

func (s *SaramaMQ) Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error {
	s.consumerMu.Lock()
	defer s.consumerMu.Unlock()

	if _, exists := s.consumers[topic]; exists {
		return fmt.Errorf("already subscribed to topic %s", topic)
	}

	// 设置默认选项
	options := defaultSubscribeOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if options.GroupID == "" {
		return errors.New("GroupID is required for sarama consumer")
	}

	// 创建消费者组
	consumer, err := sarama.NewConsumerGroup(s.brokers, options.GroupID, s.config)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	// 创建死信队列生产者（如果配置了死信队列）
	var dlqProducer sarama.SyncProducer
	if options.DeadLetterQueue != "" {
		dlqProducer, err = s.getProducer(options.DeadLetterQueue)
		if err != nil {
			log.Printf("Failed to create DLQ producer: %v", err)
		}
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	handlerObj := &consumerGroupHandler{
		mq:          s,
		topic:       topic,
		handler:     handler,
		options:     options,
		group:       consumer,
		ctx:         ctx,
		cancel:      cancel,
		dlqProducer: dlqProducer,
	}

	s.consumers[topic] = handlerObj

	// 启动消费者
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// 使用 for 循环而不是 for { select {} }
		for {
			// 检查上下文是否已取消
			if ctx.Err() != nil {
				return
			}

			// 消费消息
			err := consumer.Consume(ctx, []string{topic}, handlerObj)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return // 消费者已关闭
				}
				log.Printf("Consumer error: %v", err)
				time.Sleep(5 * time.Second) // 避免快速重试
			}
		}
	}()

	log.Printf("Subscribed to topic %s with group %s", topic, options.GroupID)
	return nil
}

// 实现 sarama.ConsumerGroupHandler 接口
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 使用 for range 直接遍历消息通道
	for msg := range claim.Messages() {
		// 检查上下文是否已取消
		if session.Context().Err() != nil {
			return nil
		}

		// 包装消息
		mqMsg := &Message{
			Topic:     msg.Topic,
			Body:      msg.Value,
			Timestamp: msg.Timestamp,
		}

		// 处理消息（带重试和超时控制）
		processed := h.handleMessage(session.Context(), mqMsg)

		// 根据配置标记消息（提交偏移量）
		if processed {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}

// handleMessage 处理消息（带重试和超时控制）
func (h *consumerGroupHandler) handleMessage(ctx context.Context, msg *Message) bool {
	retryCount := 0
	maxRetries := h.options.RetryPolicy.MaxRetries
	backoff := h.options.RetryPolicy.Backoff
	maxBackoff := h.options.RetryPolicy.MaxBackoff
	if maxBackoff == 0 {
		maxBackoff = 30 * time.Second // 默认最大退避时间
	}

	for {
		// 创建带超时的上下文
		var cancel context.CancelFunc
		msgCtx := ctx
		if h.options.MessageTimeout > 0 {
			msgCtx, cancel = context.WithTimeout(ctx, h.options.MessageTimeout)
			defer func() {
				if cancel != nil {
					cancel()
				}
			}()
		}

		// 执行消息处理
		startTime := time.Now()
		err := h.handler(msgCtx, msg)
		duration := time.Since(startTime)

		if err == nil {
			log.Printf("Message processed successfully in %v", duration)
			return true // 处理成功
		}

		// 记录错误
		log.Printf("Error processing message (attempt %d/%d, topic %s): %v",
			retryCount+1, maxRetries, msg.Topic, err)

		// 检查是否达到最大重试次数
		if maxRetries > 0 && retryCount >= maxRetries {
			log.Printf("Max retries exceeded for message on topic %s", msg.Topic)

			// 发送到死信队列
			if h.dlqProducer != nil {
				if err := h.sendToDLQ(msg); err != nil {
					log.Printf("Failed to send to DLQ: %v", err)
				} else {
					log.Printf("Message sent to DLQ: %s", h.options.DeadLetterQueue)
				}
			}
			return true // 标记为已处理（即使失败）
		}

		retryCount++

		// 计算指数退避时间
		sleepDuration := time.Duration(retryCount) * backoff
		if sleepDuration > maxBackoff {
			sleepDuration = maxBackoff
		}

		// 等待退避时间或上下文取消
		select {
		case <-ctx.Done():
			return false // 上下文取消，不标记消息
		case <-time.After(sleepDuration):
			// 继续重试
		}
	}
}

// sendToDLQ 发送消息到死信队列
func (h *consumerGroupHandler) sendToDLQ(msg *Message) error {
	dlqMsg := &sarama.ProducerMessage{
		Topic: h.options.DeadLetterQueue,
		Value: sarama.ByteEncoder(msg.Body),
		Headers: []sarama.RecordHeader{
			{Key: []byte("Original-Topic"), Value: []byte(msg.Topic)},
			{Key: []byte("Retry-Count"), Value: []byte(fmt.Sprintf("%d", h.options.RetryPolicy.MaxRetries))},
		},
	}

	_, _, err := h.dlqProducer.SendMessage(dlqMsg)
	return err
}

func (s *SaramaMQ) Unsubscribe(topic string) error {
	s.consumerMu.Lock()
	defer s.consumerMu.Unlock()

	handler, exists := s.consumers[topic]
	if !exists {
		return fmt.Errorf("not subscribed to topic %s", topic)
	}

	// 取消上下文
	handler.cancel()

	// 关闭消费者组
	if err := handler.group.Close(); err != nil {
		log.Printf("Error closing consumer group for topic %s: %v", topic, err)
	}

	delete(s.consumers, topic)
	log.Printf("Unsubscribed from topic %s", topic)
	return nil
}

func (s *SaramaMQ) CreateTopic(ctx context.Context, topic string, partitions int) error {
	detail := &sarama.TopicDetail{
		NumPartitions:     int32(partitions),
		ReplicationFactor: 1, // 生产环境应设置为3
	}

	err := s.admin.CreateTopic(topic, detail, false)
	if err != nil {
		if topicErr, ok := err.(*sarama.TopicError); ok && topicErr.Err == sarama.ErrTopicAlreadyExists {
			log.Printf("Topic %s already exists", topic)
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	log.Printf("Created topic %s with %d partitions", topic, partitions)
	return nil
}

func (s *SaramaMQ) Close() error {
	// 关闭所有消费者
	for topic := range s.consumers {
		if err := s.Unsubscribe(topic); err != nil {
			log.Printf("Error unsubscribing from topic %s: %v", topic, err)
		}
	}

	// 关闭所有生产者
	for _, producer := range s.producers {
		if err := producer.Close(); err != nil {
			log.Printf("Error closing producer: %v", err)
		}
	}

	// 关闭管理客户端
	if err := s.admin.Close(); err != nil {
		log.Printf("Error closing admin client: %v", err)
	}

	// 等待所有协程完成
	s.wg.Wait()
	log.Println("SaramaMQ closed")
	return nil
}
