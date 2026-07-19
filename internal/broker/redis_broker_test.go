package broker

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMiniredisBroker(t *testing.T) *redisBroker {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return newRedisBroker(client)
}

func recvWithin(t *testing.T, ch <-chan []byte, d time.Duration) ([]byte, bool) {
	t.Helper()
	select {
	case p := <-ch:
		return p, true
	case <-time.After(d):
		return nil, false
	}
}

func TestRedisBroker_PublishSubscribe(t *testing.T) {
	tests := []struct {
		name          string
		subscribe     []string
		publishStream string
		wantDeliver   bool
	}{
		{
			name:          "publish to subscribed stream delivers",
			subscribe:     []string{"posts:42"},
			publishStream: "posts:42",
			wantDeliver:   true,
		},
		{
			name:          "publish to different stream does not deliver",
			subscribe:     []string{"posts:42"},
			publishStream: "posts:99",
			wantDeliver:   false,
		},
		{
			name:          "multi-stream subscription delivers on any of the streams",
			subscribe:     []string{"posts:42", "chat:general"},
			publishStream: "chat:general",
			wantDeliver:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMiniredisBroker(t)

			sub, err := b.Subscribe(t.Context(), tt.subscribe...)
			require.NoError(t, err)
			defer func() { _ = sub.Close() }()

			require.NoError(t, b.Publish(context.Background(), tt.publishStream, []byte("x")))

			got, ok := recvWithin(t, sub.PayloadCh, 500*time.Millisecond)
			if tt.wantDeliver {
				require.True(t, ok, "expected payload, timed out")
				assert.Equal(t, []byte("x"), got)
			} else {
				assert.False(t, ok, "did not expect payload, got %q", got)
			}
		})
	}
}

func TestRedisBroker_Subscribe_Error(t *testing.T) {
	tests := []struct {
		name      string
		brokerFn  func(t *testing.T) *redisBroker
		streams   []string
	}{
		{
			name:     "no streams returns error",
			brokerFn: newMiniredisBroker,
			streams:  nil,
		},
		{
			name: "connect failed returns error",
			brokerFn: func(t *testing.T) *redisBroker {
				client := redis.NewClient(&redis.Options{
					Addr:        "127.0.0.1:1",
					DialTimeout: 100 * time.Millisecond,
				})
				t.Cleanup(func() { _ = client.Close() })
				return newRedisBroker(client)
			},
			streams: []string{"s"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.brokerFn(t)
			sub, err := b.Subscribe(t.Context(), tt.streams...)
			assert.Error(t, err)
			assert.Nil(t, sub)
		})
	}
}

func TestRedisBroker_Subscription_Close(t *testing.T) {
	tests := []struct {
		name       string
		closeCalls int
	}{
		{
			name:       "close is idempotent when called multiple times",
			closeCalls: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMiniredisBroker(t)
			sub, err := b.Subscribe(t.Context(), "stream")
			require.NoError(t, err)

			for i := range tt.closeCalls {
				require.NoErrorf(t, sub.Close(), "call %d", i+1)
			}
		})
	}
}

func TestRedisBroker_Subscription_CtxCancel(t *testing.T) {
	tests := []struct {
		name    string
		streams []string
	}{
		{
			name:    "single-stream subscription is torn down on ctx cancel",
			streams: []string{"posts:42"},
		},
		{
			name:    "multi-stream subscription is torn down on ctx cancel",
			streams: []string{"posts:42", "chat:general"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMiniredisBroker(t)
			ctx, cancel := context.WithCancel(t.Context())
			sub, err := b.Subscribe(ctx, tt.streams...)
			require.NoError(t, err)

			cancel()

			assert.Eventually(t, func() bool {
				_ = b.Publish(context.Background(), tt.streams[0], []byte("x"))
				_, ok := recvWithin(t, sub.PayloadCh, 50*time.Millisecond)
				return !ok
			}, 500*time.Millisecond, 50*time.Millisecond)
		})
	}
}
