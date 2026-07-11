package turbo

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamHeader(t *testing.T) {
	w := httptest.NewRecorder()

	StreamHeader(w)

	assert.Equal(t, "text/vnd.turbo-stream.html; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestNewStreamBroker(t *testing.T) {
	tests := []struct {
		name string
		cfgs []StreamBrokerConfig
	}{
		{
			name: "no config returns default (in-process) broker",
			cfgs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker(tt.cfgs...)
			assert.NotNil(t, sb)
		})
	}
}

// stubContent is a StreamContent used to drive Broadcast tests without
// wiring in the Elm helpers or a full templ component.
type stubContent struct {
	body string
	err  error
}

func (s stubContent) Render(_ context.Context, w io.Writer) error {
	if s.err != nil {
		return s.err
	}
	_, err := io.WriteString(w, s.body)
	return err
}

func TestBroadcast(t *testing.T) {
	tests := []struct {
		name        string
		contents    []StreamContent
		wantPayload []byte // nil means no delivery expected
	}{
		{
			name:        "single content is published as-is",
			contents:    []StreamContent{stubContent{body: "hello"}},
			wantPayload: []byte("hello"),
		},
		{
			name: "multiple contents are concatenated in order",
			contents: []StreamContent{
				stubContent{body: "a"},
				stubContent{body: "b"},
				stubContent{body: "c"},
			},
			wantPayload: []byte("abc"),
		},
		{
			name:     "empty contents is a no-op that returns nil",
			contents: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			sub, err := sb.Subscribe(t.Context(), "stream")
			require.NoError(t, err)
			defer sub.Close()

			require.NoError(t, Broadcast(t.Context(), sb, "stream", tt.contents...))

			select {
			case got := <-sub.PayloadCh:
				if tt.wantPayload == nil {
					t.Fatalf("unexpected delivery: %s", got)
				}
				assert.Equal(t, tt.wantPayload, got)
			default:
				if tt.wantPayload != nil {
					t.Fatal("expected delivery, got none")
				}
			}
		})
	}
}

func TestBroadcast_Error(t *testing.T) {
	errRender := errors.New("render fail")

	tests := []struct {
		name     string
		ctxFn    func(t *testing.T) context.Context
		contents []StreamContent
		wantErr  error
	}{
		{
			name:  "render error aborts before publish and returns unwrapped",
			ctxFn: func(t *testing.T) context.Context { return t.Context() },
			contents: []StreamContent{
				stubContent{body: "a"},
				stubContent{err: errRender},
				stubContent{body: "b"},
			},
			wantErr: errRender,
		},
		{
			name: "canceled ctx surfaces the ctx error from Publish",
			ctxFn: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			contents: []StreamContent{stubContent{body: "x"}},
			wantErr:  context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			sub, err := sb.Subscribe(t.Context(), "stream")
			require.NoError(t, err)
			defer sub.Close()

			err = Broadcast(tt.ctxFn(t), sb, "stream", tt.contents...)
			require.ErrorIs(t, err, tt.wantErr)

			select {
			case got := <-sub.PayloadCh:
				t.Fatalf("unexpected delivery on error: %s", got)
			default:
			}
		})
	}
}
