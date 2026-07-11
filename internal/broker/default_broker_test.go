package broker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultBroker_PublishSubscribe(t *testing.T) {
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
			b := newDefaultBroker()

			sub, err := b.Subscribe(t.Context(), tt.subscribe...)
			require.NoError(t, err)
			defer func() { _ = sub.Close() }()

			require.NoError(t, b.Publish(context.Background(), tt.publishStream, []byte("x")))

			got := drainPayloads(sub.PayloadCh)
			if tt.wantDeliver {
				assert.Len(t, got, 1)
			} else {
				assert.Empty(t, got)
			}
		})
	}
}

func TestDefaultBroker_FanOut(t *testing.T) {
	tests := []struct {
		name    string
		numSubs int
	}{
		{
			name:    "publish reaches every subscriber on the same stream",
			numSubs: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newDefaultBroker()

			subs := make([]*Subscription, tt.numSubs)
			for i := range subs {
				s, err := b.Subscribe(t.Context(), "stream")
				require.NoError(t, err)
				subs[i] = s
				defer func() { _ = s.Close() }()
			}

			require.NoError(t, b.Publish(context.Background(), "stream", []byte("fan")))

			for i, s := range subs {
				got := drainPayloads(s.PayloadCh)
				assert.Lenf(t, got, 1, "subscriber %d", i)
			}
		})
	}
}

func TestDefaultBroker_BufferOverflow(t *testing.T) {
	tests := []struct {
		name         string
		publishCount int
		wantReceived int
	}{
		{
			name:         "publishes beyond subscriber buffer are dropped",
			publishCount: subscriberBufferSize + 5,
			wantReceived: subscriberBufferSize,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newDefaultBroker()

			sub, err := b.Subscribe(t.Context(), "stream")
			require.NoError(t, err)
			defer func() { _ = sub.Close() }()

			for i := range tt.publishCount {
				require.NoError(t, b.Publish(context.Background(), "stream", []byte{byte(i)}))
			}

			got := drainPayloads(sub.PayloadCh)
			assert.Len(t, got, tt.wantReceived)
		})
	}
}

func TestDefaultBroker_Publish_Error(t *testing.T) {
	tests := []struct {
		name    string
		ctxFn   func(t *testing.T) context.Context
		wantErr error
	}{
		{
			name: "canceled ctx returns ctx error",
			ctxFn: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newDefaultBroker()
			err := b.Publish(tt.ctxFn(t), "stream", []byte("x"))
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDefaultBroker_Subscribe_Error(t *testing.T) {
	tests := []struct {
		name    string
		streams []string
	}{
		{
			name:    "no streams returns error",
			streams: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newDefaultBroker()
			sub, err := b.Subscribe(t.Context(), tt.streams...)
			assert.Error(t, err)
			assert.Nil(t, sub)
		})
	}
}

func TestDefaultBroker_Subscription_Close(t *testing.T) {
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
			b := newDefaultBroker()
			sub, err := b.Subscribe(t.Context(), "stream")
			require.NoError(t, err)

			for i := range tt.closeCalls {
				require.NoErrorf(t, sub.Close(), "call %d", i+1)
			}
		})
	}
}

func TestDefaultBroker_Subscription_CtxCancel(t *testing.T) {
	tests := []struct {
		name    string
		streams []string
	}{
		{
			name:    "single-stream subscription is unregistered on ctx cancel",
			streams: []string{"posts:42"},
		},
		{
			name:    "multi-stream subscription is unregistered from every stream on ctx cancel",
			streams: []string{"posts:42", "chat:general"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newDefaultBroker()
			ctx, cancel := context.WithCancel(t.Context())
			_, err := b.Subscribe(ctx, tt.streams...)
			require.NoError(t, err)

			cancel()

			assert.Eventually(t, func() bool {
				b.mu.RLock()
				defer b.mu.RUnlock()
				for _, s := range tt.streams {
					if _, ok := b.subscribers[s]; ok {
						return false
					}
				}
				return true
			}, 200*time.Millisecond, 5*time.Millisecond)
		})
	}
}

// drainPayloads reads everything currently buffered in ch without blocking.
func drainPayloads(ch <-chan []byte) [][]byte {
	var got [][]byte
	for {
		select {
		case p := <-ch:
			got = append(got, p)
		default:
			return got
		}
	}
}
