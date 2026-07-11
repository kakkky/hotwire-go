package broker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

const (
	subscriberBufferSize = 32
)

type defaultBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []byte]struct{}
}

func newDefaultBroker() *defaultBroker {
	return &defaultBroker{
		subscribers: make(map[string]map[chan []byte]struct{}),
	}
}

func (b *defaultBroker) Publish(ctx context.Context, stream string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.RLock()
	targetStream := make([]chan []byte, 0, len(b.subscribers[stream]))
	for ch := range b.subscribers[stream] {
		targetStream = append(targetStream, ch)
	}
	b.mu.RUnlock()

	for _, subscriber := range targetStream {
		select {
		case subscriber <- payload:
		default:
			slog.Warn("broker: dropped payload", "stream", stream)
		}
	}

	return nil
}

func (b *defaultBroker) Subscribe(ctx context.Context, streams ...string) (*Subscription, error) {
	if len(streams) == 0 {
		return nil, errors.New("broker: at least one stream required")
	}

	ch := make(chan []byte, subscriberBufferSize)
	b.mu.Lock()
	for _, s := range streams {
		if b.subscribers[s] == nil {
			b.subscribers[s] = make(map[chan []byte]struct{})
		}
		b.subscribers[s][ch] = struct{}{}
	}
	b.mu.Unlock()

	var once sync.Once
	unSubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			for _, s := range streams {
				delete(b.subscribers[s], ch)
				if len(b.subscribers[s]) == 0 {
					delete(b.subscribers, s)
				}
			}
			b.mu.Unlock()
		})
	}

	stopCtxWatch := context.AfterFunc(ctx, unSubscribe)

	return &Subscription{
		PayloadCh: ch,
		unSubscribe: func() error {
			stopCtxWatch()
			unSubscribe()
			return nil
		},
	}, nil
}
