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
	"time"

	"github.com/kakkky/hotwire-go/internal/auth"
)

// StreamsSSEPath is the URL path served by StreamSSEHandler and
// embedded into the src attribute produced by StreamSourceSSE.
// Applications that expose the SSE endpoint at a different path
// should reassign this variable at start-up, before any handler is
// mounted or any StreamSourceSSE renders:
//
//	turbo.StreamsSSEPath = "/..."
//	mux.Handle(turbo.StreamsSSEPath, turbo.StreamSSEHandler(sb))
var StreamsSSEPath = "/turbo_streams_sse"

// StreamSourceSSE renders a <turbo-stream-source> that subscribes
// the current page to the named broker stream over SSE. The signed
// token in its src binds the stream to the caller's session sid, so
// a leaked token cannot be replayed from another browser.
//
// ctx must carry a sid installed by StreamsMiddleware; otherwise
// this function panics.
//
// Place the element inside <body>: <turbo-stream-source> disconnects
// when it leaves the document, and Turbo Drive replaces the body on
// every full-page navigation.
//
// Register via turbo.TemplateFuncMap and call from a template:
//
//	{{ turboStreamSourceSSE .Ctx "..." }}
//
// Or from an a-h/templ template (https://github.com/a-h/templ):
//
//	@turbo.StreamSourceSSE(ctx, "...")
func StreamSourceSSE(ctx context.Context, stream string) Elm {
	sid, ok := ctx.Value(sidCtxKey{}).(string)
	if !ok {
		panic("turbo: StreamSourceSSE requires turbo.StreamsMiddleware in the request pipeline")
	}
	token := auth.SignToken(stream, sid, defaultStreamTokenTTL)
	q := url.Values{}
	q.Set("token", token)
	u := url.URL{
		Path:     StreamsSSEPath,
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

// StreamSSEHandler returns the SSE endpoint that pairs with
// StreamSourceSSE: it accepts the EventSource opened by a
// <turbo-stream-source>, verifies the request, and forwards every
// payload Broadcast publishes to the requested stream on sb through
// to the client. Unauthorized requests return 401 without opening a
// subscription.
//
// Wire it under StreamsSSEPath. StreamsMiddleware wraps only the
// page routes; this handler reads the session cookie directly:
//
//	sb := turbo.NewStreamBroker()
//	mux.Handle(turbo.StreamsSSEPath, turbo.StreamSSEHandler(sb))
//	mux.Handle("/", turbo.StreamsMiddleware(pageHandler))
//
// Heartbeat interval is configurable via WithHeartbeatInterval
// (default 15s).
func StreamSSEHandler(sb StreamBroker, cfgs ...StreamConfig) http.Handler {
	c := &streamConfigs{
		heartbeatInterval: defaultHeartbeatInterval,
	}
	for _, cfg := range cfgs {
		cfg(c)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := checkSameOrigin(r); err != nil {
			slog.Error("turbo: SSE request rejected", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := r.URL.Query().Get("token")

		stream, sid, err := auth.VerifyToken(token)
		if err != nil {
			slog.Error("turbo: SSE token verification failed", "token", token, "error", err)
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
			slog.Error("turbo: SSE session ID mismatch", "token", token)
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

		h := w.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("Connection", "keep-alive")
		h.Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			slog.Error("turbo: response writer does not support flushing")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		flusher.Flush()

		heartbeat := time.NewTicker(c.heartbeatInterval)
		defer heartbeat.Stop()

		for {
			select {
			case payload := <-subscription.PayloadCh:
				if err := writeEventStreamFormat(w, "message", payload, 3*time.Second); err != nil {
					return
				}
				flusher.Flush()
			case <-heartbeat.C:
				if _, err := io.WriteString(w, ":\n\n"); err != nil {
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})
}

// writeEventStreamFormat writes one SSE event to w as an optional
// "event:" line, an optional "retry:" line, one "data:" line per newline
// segment of data, and a terminating blank line.
//
// Per the HTML Living Standard the SSE parser terminates a field on any
// of "\n", "\r", or "\r\n". A bare "\r" inside data would let subsequent
// bytes be interpreted as new SSE fields, so both "\r\n" and bare "\r"
// are normalized to "\n" before splitting to prevent event injection.
func writeEventStreamFormat(w io.Writer, event string, data []byte, retry time.Duration) error {
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}

	if retry > 0 {
		if _, err := fmt.Fprintf(w, "retry: %d\n", retry.Milliseconds()); err != nil {
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

func checkSameOrigin(r *http.Request) error {
	if origin := r.Header.Get("Origin"); origin != "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if origin != scheme+"://"+r.Host {
			return errors.New("turbo: cross-origin request rejected")
		}
	}
	return nil
}
