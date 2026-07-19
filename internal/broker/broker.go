package broker

import (
	"context"
)

// Broker is a pub/sub hub keyed by named streams. Publish fans a byte
// payload out to every current subscriber of the given stream, and
// Subscribe returns a handle that receives payloads for one or more
// streams over a channel. The interface is intentionally
// backend-agnostic so callers can swap the in-process default for a
// distributed backend without touching call sites.
//
// Implementations are expected to be safe for concurrent use: a single
// Broker is meant to be constructed once at start-up via New and shared
// across handlers and goroutines. Publish must not block on slow
// subscribers, and Subscribe must honor context cancellation so a
// consumer that walks away does not leak resources on the broker side.
type Broker interface {
	Publish(ctx context.Context, stream string, payload []byte) error
	Subscribe(ctx context.Context, streams ...string) (*Subscription, error)
}

// Subscription is the handle Broker.Subscribe returns. Received
// payloads arrive on PayloadCh in publish order per stream; when the
// caller subscribes to more than one stream, payloads from all of them
// are multiplexed onto the same channel.
//
// The subscription stays live until either Close is called explicitly
// or the context passed to Subscribe is canceled — whichever happens
// first. After teardown, PayloadCh is no longer written to; callers
// that select on it should also select on ctx.Done to exit cleanly.
type Subscription struct {
	PayloadCh <-chan []byte

	unSubscribe func() error
}

// Close tears down the subscription: it removes the underlying channel
// from every stream's subscriber set and stops the context watcher
// started by Subscribe. It is safe to call more than once and safe to
// call concurrently with the context-driven teardown — the first
// caller wins and later calls become no-ops. The returned error is
// currently always nil; the signature reserves room for backends
// whose teardown can fail (for example a Redis client returning an
// I/O error while unsubscribing).
func (s *Subscription) Close() error {
	return s.unSubscribe()
}

// New constructs a Broker, choosing its backend from the given
// Configs. With no Configs the returned value is backed by an
// in-process implementation: a single Go process, no persistence, no
// fan-out across replicas. Passing WithRedis switches the backend to
// Redis PUB/SUB so multiple processes can share the same stream
// namespace.
//
// The returned Broker is safe for concurrent use and is intended to
// be built once at application start-up and shared across handlers.
func New(cfgs ...Config) Broker {
	c := &configs{}
	for _, cfg := range cfgs {
		cfg(c)
	}
	var broker Broker
	switch {
	case c.redisClient != nil:
		broker = newRedisBroker(c.redisClient)
	default:
		broker = newDefaultBroker()
	}
	return broker
}
