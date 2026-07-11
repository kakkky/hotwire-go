package turbo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsStreamRequest(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{
			name:   "Accept contains turbo-stream MIME",
			accept: "text/vnd.turbo-stream.html, text/html",
			want:   true,
		},
		{
			name:   "Accept is only turbo-stream MIME",
			accept: "text/vnd.turbo-stream.html",
			want:   true,
		},
		{
			name:   "Accept is only text/html",
			accept: "text/html",
			want:   false,
		},
		{
			name:   "Accept header missing",
			accept: "",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.accept != "" {
				r.Header.Set("Accept", tt.accept)
			}
			assert.Equal(t, tt.want, IsStreamRequest(r))
		})
	}
}

func TestRequestID(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "header value is returned as-is",
			value: "01FAKE-UUID",
			want:  "01FAKE-UUID",
		},
		{
			name:  "missing header returns empty",
			value: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.value != "" {
				r.Header.Set("X-Turbo-Request-Id", tt.value)
			}
			assert.Equal(t, tt.want, RequestID(r))
		})
	}
}

func TestStreamHeader(t *testing.T) {
	tests := []struct {
		name            string
		wantContentType string
	}{
		{
			name:            "sets the Turbo Streams MIME type as Content-Type",
			wantContentType: "text/vnd.turbo-stream.html; charset=utf-8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			StreamHeader(w)
			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
		})
	}
}

func TestStreamActionHelpers(t *testing.T) {
	// Each variant is called once with just a target and once with an
	// extra Attrs bundle so both the "target != ''" and the extra-attrs
	// loop are exercised for every action helper.
	tests := []struct {
		name       string
		got        Elm
		action     string
		wantTarget string
		wantExtra  Attrs
	}{
		{name: "StreamAppend/target-only", got: StreamAppend("messages"), action: "append", wantTarget: "messages"},
		{name: "StreamAppend/with-extra", got: StreamAppend("messages", AttrMethodMorph()), action: "append", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
		{name: "StreamPrepend/target-only", got: StreamPrepend("messages"), action: "prepend", wantTarget: "messages"},
		{name: "StreamPrepend/with-extra", got: StreamPrepend("messages", AttrMethodMorph()), action: "prepend", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
		{name: "StreamReplace/target-only", got: StreamReplace("messages"), action: "replace", wantTarget: "messages"},
		{name: "StreamReplace/with-extra", got: StreamReplace("messages", AttrMethodMorph()), action: "replace", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
		{name: "StreamUpdate/target-only", got: StreamUpdate("messages"), action: "update", wantTarget: "messages"},
		{name: "StreamUpdate/with-extra", got: StreamUpdate("messages", AttrMethodMorph()), action: "update", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
		{name: "StreamRemove/target-only", got: StreamRemove("messages"), action: "remove", wantTarget: "messages"},
		{name: "StreamRemove/with-extra", got: StreamRemove("messages", AttrMethodMorph()), action: "remove", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
		{name: "StreamBefore/target-only", got: StreamBefore("messages"), action: "before", wantTarget: "messages"},
		{name: "StreamBefore/with-extra", got: StreamBefore("messages", AttrMethodMorph()), action: "before", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
		{name: "StreamAfter/target-only", got: StreamAfter("messages"), action: "after", wantTarget: "messages"},
		{name: "StreamAfter/with-extra", got: StreamAfter("messages", AttrMethodMorph()), action: "after", wantTarget: "messages", wantExtra: Attrs{{Key: "method", Value: "morph"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, Tag("turbo-stream"), tt.got.Tag)
			assert.Equal(t, Tag("template"), tt.got.InnerTag)
			want := Attrs{
				{Key: "action", Value: tt.action},
				{Key: "target", Value: tt.wantTarget},
			}
			want = append(want, tt.wantExtra...)
			assert.Equal(t, want, tt.got.Attrs)
		})
	}
}

func TestStreamActionHelpers_EmptyTargetOmitsTargetAttr(t *testing.T) {
	// An empty target is meaningful with AttrTargets (multi-target streams):
	// the action helper drops the target attribute so only targets is emitted.
	tests := []struct {
		name string
		got  Elm
		want Attrs
	}{
		{
			name: "StreamRemove with empty target and AttrTargets",
			got:  StreamRemove("", AttrTargets(".notification")),
			want: Attrs{
				{Key: "action", Value: "remove"},
				{Key: "targets", Value: ".notification"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got.Attrs)
		})
	}
}

func TestStreamRefresh(t *testing.T) {
	tests := []struct {
		name  string
		extra []Attrs
		want  Attrs
	}{
		{
			name: "refresh has no target",
			want: Attrs{{Key: "action", Value: "refresh"}},
		},
		{
			name:  "refresh accepts request-id for dedup",
			extra: []Attrs{AttrRequestID("req-1")},
			want: Attrs{
				{Key: "action", Value: "refresh"},
				{Key: "request-id", Value: "req-1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StreamRefresh(tt.extra...)
			assert.Equal(t, Tag("turbo-stream"), got.Tag)
			assert.Equal(t, Tag("template"), got.InnerTag)
			assert.Equal(t, tt.want, got.Attrs)
		})
	}
}

func TestStreamAttrHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  Attrs
		want Attrs
	}{
		{
			name: "AttrTargets",
			got:  AttrTargets(".notification"),
			want: Attrs{{Key: "targets", Value: ".notification"}},
		},
		{
			name: "AttrRequestID",
			got:  AttrRequestID("req-42"),
			want: Attrs{{Key: "request-id", Value: "req-42"}},
		},
		{
			name: "AttrMethodMorph",
			got:  AttrMethodMorph(),
			want: Attrs{{Key: "method", Value: "morph"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
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
