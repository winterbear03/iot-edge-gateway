package bus

import "context"

type DataBus interface {
    Publish(ctx context.Context, channel string, data []byte) error
    Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
}
