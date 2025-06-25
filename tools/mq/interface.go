package mq

import (
	"context"
	"time"
)

type HandlerFunc func(context.Context, *Message) error

type MQ interface {
	// 生产者接口
	Publish(ctx context.Context, msg *Message) error
	BatchPublish(ctx context.Context, msgs []*Message) error

	// 消费者接口
	Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error
	Unsubscribe(topic string) error

	// 管理接口
	CreateTopic(ctx context.Context, topic string, partitions int) error
	Close() error
}

type Message struct {
	Topic     string
	Body      []byte
	Key       string
	Timestamp time.Time
}

// SubscribeOption 是用于配置订阅选项的函数类型
type SubscribeOption func(*SubscribeOptions)

// WithGroupID 设置消费者组ID
func WithGroupID(groupID string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.GroupID = groupID
	}
}

// WithAutoAck 设置是否自动确认消息
func WithAutoAck(autoAck bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = autoAck
	}
}

// WithConcurrency 设置并发处理消息的协程数量
func WithConcurrency(concurrency int) SubscribeOption {
	return func(o *SubscribeOptions) {
		if concurrency < 1 {
			// 确保并发数至少为1
			o.Concurrency = 1
		} else {
			o.Concurrency = concurrency
		}
	}
}

// WithRetryPolicy 设置消息处理失败的重试策略
func WithRetryPolicy(maxRetries int, backoff time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		if maxRetries < 0 {
			// 确保重试次数非负
			o.RetryPolicy.MaxRetries = 0
		} else {
			o.RetryPolicy.MaxRetries = maxRetries
		}

		if backoff < 0 {
			// 确保退避时间为非负
			o.RetryPolicy.Backoff = 0
		} else {
			o.RetryPolicy.Backoff = backoff
		}
	}
}

// WithMaxBackoff 设置最大退避时间（指数退避上限）
func WithMaxBackoff(maxBackoff time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		if maxBackoff < 0 {
			o.RetryPolicy.MaxBackoff = 0
		} else {
			o.RetryPolicy.MaxBackoff = maxBackoff
		}
	}
}

// WithDeadLetterQueue 设置死信队列选项
func WithDeadLetterQueue(topic string, maxRetries int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.DeadLetterQueue = topic
		if maxRetries > 0 {
			o.RetryPolicy.MaxRetries = maxRetries
		}
	}
}

// WithOffsetInitial 设置初始消费偏移量
func WithOffsetInitial(offset OffsetInitial) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.OffsetInitial = offset
	}
}

// WithBatchSize 设置批量处理大小
func WithBatchSize(batchSize int) SubscribeOption {
	return func(o *SubscribeOptions) {
		if batchSize < 1 {
			o.BatchSize = 1
		} else {
			o.BatchSize = batchSize
		}
	}
}

// WithMessageTimeout 设置消息处理超时时间
func WithMessageTimeout(timeout time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		if timeout < 0 {
			o.MessageTimeout = 0
		} else {
			o.MessageTimeout = timeout
		}
	}
}

// 扩展的 SubscribeOptions 结构体
type SubscribeOptions struct {
	GroupID         string        // 消费者组ID
	AutoAck         bool          // 是否自动确认消息
	Concurrency     int           // 并发处理协程数量
	RetryPolicy     RetryPolicy   // 重试策略
	DeadLetterQueue string        // 死信队列主题
	OffsetInitial   OffsetInitial // 初始消费偏移量
	BatchSize       int           // 批量处理大小
	MessageTimeout  time.Duration // 消息处理超时时间
}

// RetryPolicy 扩展版本
type RetryPolicy struct {
	MaxRetries int           // 最大重试次数
	Backoff    time.Duration // 基础退避时间
	MaxBackoff time.Duration // 最大退避时间（指数退避上限）
}

// OffsetInitial 定义初始偏移量类型
type OffsetInitial int

const (
	OffsetNewest OffsetInitial = iota // 从最新偏移量开始消费
	OffsetOldest                      // 从最早偏移量开始消费
	OffsetStored                      // 从存储的偏移量开始消费
)

// defaultSubscribeOptions 提供默认订阅选项
func defaultSubscribeOptions() SubscribeOptions {
	return SubscribeOptions{
		AutoAck:     true,
		Concurrency: 1,
		RetryPolicy: RetryPolicy{
			MaxRetries: 3,
			Backoff:    1 * time.Second,
			MaxBackoff: 30 * time.Second,
		},
		BatchSize:      1,
		MessageTimeout: 30 * time.Second,
		OffsetInitial:  OffsetNewest,
	}
}
