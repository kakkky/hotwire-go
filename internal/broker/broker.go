package broker

import (
	"context"
)

type Broker interface {
	Publish(ctx context.Context, stream string, payload []byte) error
	Subscribe(ctx context.Context, streams ...string) (*Subscription, error)
}

type Subscription struct {
	PayloadCh <-chan []byte

	unSubscribe func() error
}

func (s *Subscription) Close() error {
	return s.unSubscribe()
}

func New(cfgs ...Config) Broker {
	c := &configs{}
	for _, cfg := range cfgs {
		cfg(c)
	}
	var broker Broker
	switch {
	case c.redisClient != nil:

	default:
		broker = newDefaultBroker()
	}
	return broker
}
