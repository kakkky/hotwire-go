package broker

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
)

type redisBroker struct {
	client redis.UniversalClient
}

func newRedisBroker(client redis.UniversalClient) *redisBroker {
	return &redisBroker{client: client}
}

func (b *redisBroker) Publish(ctx context.Context, stream string, payload []byte) error {
	return b.client.Publish(ctx, stream, payload).Err()
}

func (b *redisBroker) Subscribe(ctx context.Context, streams ...string) (*Subscription, error) {
	if len(streams) == 0 {
		return nil, errors.New("broker: at least one stream required")
	}

	ps := b.client.Subscribe(ctx, streams...)
	if err := ps.Ping(ctx); err != nil {
		_ = ps.Close()
		return nil, err
	}

	ch := make(chan []byte, subscriberBufferSize)

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			_ = ps.Close()
		})
	}
	stopCtxWatch := context.AfterFunc(ctx, unsubscribe)

	go func() {
		for msg := range ps.Channel() {
			select {
			case ch <- []byte(msg.Payload):
			default:
				slog.Warn("broker: dropped payload", "stream", msg.Channel)
			}
		}
	}()

	return &Subscription{
		PayloadCh: ch,
		unSubscribe: func() error {
			stopCtxWatch()
			unsubscribe()
			return nil
		},
	}, nil
}
