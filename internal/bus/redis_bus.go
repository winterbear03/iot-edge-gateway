package bus

import (
    "context"
    "github.com/go-redis/redis/v8"
)

type RedisBus struct { client *redis.Client }

func NewRedisBus(addr string) *RedisBus {
    rdb := redis.NewClient(&redis.Options{Addr: addr})
    return &RedisBus{client: rdb}
}

func (r *RedisBus) Publish(ctx context.Context, channel string, data []byte) error {
    return r.client.Publish(ctx, channel, data).Err()
}

func (r *RedisBus) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
    pubsub := r.client.Subscribe(ctx, channel)
    ch := make(chan []byte)
    go func() {
        defer pubsub.Close()
        for msg := range pubsub.Channel() { ch <- []byte(msg.Payload) }
    }()
    return ch, nil
}
