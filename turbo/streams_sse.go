package turbo

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/kakkky/hotwire-go/internal/auth"
)

// StreamSSEBroadcastPath is the URL path served by
// StreamSSEBroadcastHandler and embedded into the src attribute
// produced by StreamSSEBroadcastSource. Applications that expose the
// SSE endpoint at a different path should reassign this variable at
// start-up, before any handler is mounted or any
// StreamSSEBroadcastSource renders:
//
//	turbo.StreamSSEBroadcastPath = "/..."
//	mux.Handle(turbo.StreamSSEBroadcastPath, turbo.StreamSSEBroadcastHandler(sb))
var StreamSSEBroadcastPath = "/turbo_streams_sse"

// ErrStreamSSEFlushNotSupported is returned by NewStreamSSEWriter
// when the http.ResponseWriter cannot flush, so events would sit in a
// buffer instead of reaching the client as they are written.
var ErrStreamSSEFlushNotSupported = errors.New("turbo: stream SSE requires a flushable response writer")

// StreamSSEBroadcastSource renders a <turbo-stream-source> that subscribes
// the current page to the named broker stream over SSE. The signed
// token in its src binds the stream to the caller's session sid, so
// a leaked token cannot be replayed from another browser.
//
// ctx must carry a sid installed by StreamSSEBroadcastMiddleware; otherwise
// this function panics.
//
// Place the element inside <body>: <turbo-stream-source> disconnects
// when it leaves the document, and Turbo Drive replaces the body on
// every full-page navigation.
//
// Register via turbo.TemplateFuncMap and call from a template:
//
//	{{ turboStreamSSEBroadcastSource .Ctx "..." }}
//
// Or from an a-h/templ template (https://github.com/a-h/templ):
//
//	@turbo.StreamSSEBroadcastSource(ctx, "...")
func StreamSSEBroadcastSource(ctx context.Context, stream string) Elm {
	sid, ok := ctx.Value(sidCtxKey{}).(string)
	if !ok {
		panic("turbo: StreamSSEBroadcastSource requires turbo.StreamSSEBroadcastMiddleware in the request pipeline")
	}
	token := auth.SignToken(stream, sid, defaultStreamTokenTTL)
	q := url.Values{}
	q.Set("token", token)
	u := url.URL{
		Path:     StreamSSEBroadcastPath,
		RawQuery: q.Encode(),
	}
	return Elm{
		Tag: Tag("turbo-stream-source"),
		Attrs: Attrs{{
			Key:   "src",
			Value: u.String(),
		}},
	}
}

// StreamSSEWriter writes Turbo Streams as Server-Sent Events onto the
// single connection behind one http.ResponseWriter.
//
// EventSource reconnects whenever the server closes the connection,
// so finish the stream by writing a StreamRemove that removes the
// <turbo-stream-source> element.
//
// Write and Heartbeat are safe to call from multiple goroutines.
type StreamSSEWriter struct {
	mu sync.Mutex
	w  io.Writer
	rc *http.ResponseController
}

// NewStreamSSEWriter sets the text/event-stream response headers on w
// and flushes them, committing a 200 response. It returns
// ErrStreamSSEFlushNotSupported without touching w when w cannot
// flush.
//
// Flushing goes through http.ResponseController, so when w wraps
// another http.ResponseWriter, w can be flushed as long as it exposes
// the wrapped writer from an Unwrap() http.ResponseWriter method.
//
//	sse, err := turbo.NewStreamSSEWriter(w)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	_ = sse.Write(r.Context(), turbo.StreamUpdate("answer"))
func NewStreamSSEWriter(w http.ResponseWriter) (*StreamSSEWriter, error) {
	headers := map[string]string{
		"Content-Type":      "text/event-stream",
		"Cache-Control":     "no-cache",
		"Connection":        "keep-alive",
		"X-Accel-Buffering": "no",
	}
	h := w.Header()
	for k, v := range headers {
		h.Set(k, v)
	}
	rc := http.NewResponseController(w)
	if err := rc.Flush(); err != nil {
		if !errors.Is(err, http.ErrNotSupported) {
			return nil, err
		}
		// ResponseController reports a missing Flush only by trying it,
		// so the headers set above are taken back to leave w untouched.
		for k := range headers {
			h.Del(k)
		}
		return nil, ErrStreamSSEFlushNotSupported
	}
	return &StreamSSEWriter{w: w, rc: rc}, nil
}

// Write renders contents in order and sends their concatenation as a
// single SSE event. If any Render returns an error, Write returns it
// without writing anything.
func (s *StreamSSEWriter) Write(ctx context.Context, contents ...StreamContent) error {
	if len(contents) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for _, c := range contents {
		if err := c.Render(ctx, &buf); err != nil {
			return err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := writeEventStreamFormat(s.w, "message", buf.Bytes()); err != nil {
		return err
	}
	return s.rc.Flush()
}

// Heartbeat writes an SSE comment frame (":\n\n") so that reverse
// proxies do not close an idle connection.
func (s *StreamSSEWriter) Heartbeat() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := io.WriteString(s.w, ":\n\n"); err != nil {
		return err
	}
	return s.rc.Flush()
}

// StreamSSEBroadcastHandler returns the SSE endpoint that pairs with
// StreamSSEBroadcastSource: it accepts the EventSource opened by a
// <turbo-stream-source>, verifies the request, and forwards every
// payload Broadcast publishes to the requested stream on sb through
// to the client. Unauthorized requests return 401 without opening a
// subscription.
//
// Wire it under StreamSSEBroadcastPath. StreamSSEBroadcastMiddleware
// wraps only the page routes; this handler reads the session cookie
// directly:
//
//	sb := turbo.NewStreamBroker()
//	mux.Handle(turbo.StreamSSEBroadcastPath, turbo.StreamSSEBroadcastHandler(sb))
//	mux.Handle("/", turbo.StreamSSEBroadcastMiddleware(pageHandler))
//
// Heartbeat interval is configurable via
// WithStreamSSEBroadcastHeartbeatInterval (default 15s).
func StreamSSEBroadcastHandler(sb StreamBroker, cfgs ...StreamSSEBroadcastConfig) http.Handler {
	c := &streamSSEBroadcastConfigs{
		heartbeatInterval: defaultHeartbeatInterval,
	}
	for _, cfg := range cfgs {
		cfg(c)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")

		stream, sid, err := auth.VerifyToken(token)
		if err != nil {
			slog.Error("turbo: SSE token verification failed", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		cookie, err := r.Cookie(streamsSessionCookieName)
		if err != nil || cookie.Value == "" {
			slog.Error("turbo: SSE request missing session cookie", "cookie", streamsSessionCookieName)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sidFromCookie, err := auth.VerifySid(cookie.Value)
		if err != nil {
			slog.Error("turbo: SSE session cookie verification failed", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if subtle.ConstantTimeCompare([]byte(sid), []byte(sidFromCookie)) != 1 {
			slog.Error("turbo: SSE session ID mismatch")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		subscription, err := sb.Subscribe(r.Context(), stream)
		if err != nil {
			slog.Error("turbo: sse subscribe failed", "stream", stream, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() { _ = subscription.Close() }()

		sse, err := NewStreamSSEWriter(w)
		if err != nil {
			slog.Error("turbo: sse writer unavailable", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		heartbeat := time.NewTicker(c.heartbeatInterval)
		defer heartbeat.Stop()

		for {
			select {
			case payload := <-subscription.PayloadCh:
				if err := sse.Write(r.Context(), encodedStreamContent(payload)); err != nil {
					return
				}
			case <-heartbeat.C:
				if err := sse.Heartbeat(); err != nil {
					return
				}
			case <-r.Context().Done():
				return
			}
		}
	})
}

// encodedStreamContent is a Turbo Streams fragment that has already
// been rendered, such as a payload received from a StreamBroker.
type encodedStreamContent []byte

func (c encodedStreamContent) Render(_ context.Context, w io.Writer) error {
	_, err := w.Write(c)
	return err
}

// writeEventStreamFormat writes one SSE event to w as an optional
// "event:" line, one "data:" line per newline segment of data, and a
// terminating blank line.
//
// Per the HTML Living Standard the SSE parser terminates a field on any
// of "\n", "\r", or "\r\n". A bare "\r" inside data would let subsequent
// bytes be interpreted as new SSE fields, so both "\r\n" and bare "\r"
// are normalized to "\n" before splitting to prevent event injection.
func writeEventStreamFormat(w io.Writer, event string, data []byte) error {
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}

	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))
	for line := range bytes.SplitSeq(normalized, []byte("\n")) {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}
