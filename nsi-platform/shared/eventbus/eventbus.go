package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type EventHandler func(ctx context.Context, event *Event) error

type EventBus interface {
	Publish(ctx context.Context, stream string, event *Event) error
	Subscribe(ctx context.Context, stream, group, consumer string, handler EventHandler) error
	Close() error
}

type RedisBus struct {
	client *redis.Client
}

func NewRedisBus(redisURL string) (*RedisBus, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis url cannot be empty")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisBus{client: client}, nil
}

func (b *RedisBus) Publish(ctx context.Context, stream string, event *Event) error {
	if stream == "" {
		return fmt.Errorf("stream name cannot be empty")
	}
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	event.Timestamp = time.Now()

	id, err := b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"type":      event.Type,
			"payload":   string(event.Payload),
			"timestamp": event.Timestamp.UnixMilli(),
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	event.ID = id
	return nil
}

func (b *RedisBus) Subscribe(ctx context.Context, stream, group, consumer string, handler EventHandler) error {
	b.client.XGroupCreate(ctx, stream, group, "0")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			entries, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Count:    1,
				Block:    0,
			}).Result()
			if err != nil {
				return fmt.Errorf("failed to read from stream: %w", err)
			}

			for _, msg := range entries[0].Messages {
				event := &Event{
					ID:      msg.ID,
					Type:    fmt.Sprintf("%v", msg.Values["type"]),
					Payload: []byte(fmt.Sprintf("%v", msg.Values["payload"])),
				}
				if ts, ok := msg.Values["timestamp"].(int64); ok {
					event.Timestamp = time.UnixMilli(ts)
				}

				if err := handler(ctx, event); err != nil {
					return err
				}

				b.client.XAck(ctx, stream, group, msg.ID)
			}
		}
	}
}

func (b *RedisBus) Close() error {
	return b.client.Close()
}
