package turbo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/kakkky/hotwire-go/internal/auth"
)

// StreamsSSEPath is the default URL path served by StreamSSEHandler and
// embedded into the src attribute produced by StreamSourceSSE. Applications
// that need to expose the SSE endpoint at a different path should reassign
// this variable at start-up — before any handler is mounted and before any
// layout renders a StreamSourceSSE element — so both the handler and the
// generated <turbo-stream-source> tags stay in agreement:
//
//	turbo.StreamsSSEPath = "/..."
//	mux.Handle(turbo.StreamsSSEPath, turbo.StreamSSEHandler(sb))
//
// The value is a plain path, not a full URL: the browser resolves it
// against the current origin, which keeps the Origin check in
// authorizeStreamRequest simple (same-origin fetches only).
var StreamsSSEPath = "/turbo_streams_sse"

// StreamSourceSSE returns Turbo's built-in <turbo-stream-source> custom
// element with its src attribute set to StreamsSSEPath and a signed token
// carrying the given stream name. Dropping the result into a layout
// subscribes the current page to that stream over Server-Sent Events, and
// Turbo applies every <turbo-stream> fragment the server pushes.
//
// The stream name is not sent to the client verbatim: it is embedded in
// an HMAC-signed, TTL-scoped token (see signStreamToken) and only that
// token appears in the src query string. StreamSSEHandler decodes the
// token to learn which broker stream to subscribe to, so a client cannot
// rewrite the URL to eavesdrop on a stream it was not authorized for.
//
// Turbo selects the transport by inspecting the src scheme: ws:// or
// wss:// URLs open a WebSocket, anything else opens an EventSource. This
// helper emits a same-origin path (no scheme), which resolves to http(s)
// and therefore always takes the EventSource / SSE path.
//
// Place the element as a child of <body>, not <head>: <turbo-stream-source>
// disconnects when it leaves the document, and Turbo Drive replaces the
// body on every full-page navigation — a source mounted in <head> would
// survive navigation and hold a stale subscription. Register via
// turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamSourceSSE "..." }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamSourceSSE("...")
//
// Turbo Handbook — Integration with server-side frameworks:
// https://turbo.hotwired.dev/handbook/streams#integration-with-server-side-frameworks
func StreamSourceSSE(stream string) Elm {
	return Elm{
		Tag: Tag("turbo-stream-source"),
		LazyAttrs: func(ctx context.Context) Attrs {
			token := auth.SignToken(ctx, stream, defaultStreamTokenTTL)
			q := url.Values{}
			q.Set("token", token)
			u := url.URL{
				Path:     StreamsSSEPath,
				RawQuery: q.Encode(),
			}
			return Attrs{{
				Key:   "src",
				Value: u.String(),
			}}
		},
	}
}

// StreamSSEHandler returns an http.Handler that upgrades the request to a
// Server-Sent Events response and forwards every payload published to the
// requested stream on sb straight through to the client. Pair it with
// StreamSourceSSE on the render side: the layout emits a
// <turbo-stream-source> pointing at StreamsSSEPath, Turbo opens an
// EventSource against this handler, and each payload written by Broadcast
// or a direct sb.Publish reaches every currently connected subscriber as
// one SSE message.
//
// The handler pulls the target stream out of the signed token in the
// query string via authorizeStreamRequest — a missing, tampered,
// expired, or cross-origin request is answered with 401 and no
// subscription is opened. On success it sets the SSE response headers
// (text/event-stream, no-cache, keep-alive, X-Accel-Buffering: no so
// nginx does not buffer), flushes them, and enters a select loop that
// forwards payloads, emits heartbeat comments, and exits when the client
// context is canceled. Each payload is written as a single SSE event
// whose data lines mirror the payload's newline structure, so a
// multi-<turbo-stream> message stays a single event on the wire.
//
// Heartbeat comments (":\n\n") are sent at the interval configured via
// WithHeartbeatInterval (default 15s) to defeat idle-connection timeouts
// on reverse proxies and load balancers. Without them, an otherwise
// quiet stream would look dead to nginx or an ELB and get dropped.
//
// Typical wiring:
//
//	sb := turbo.NewStreamBroker()
//	mux.Handle(turbo.StreamsSSEPath, turbo.StreamSSEHandler(sb))
//
// Deployment note: the tokens minted by StreamSourceSSE are signed with
// the key held in package-scope by streams_auth.go, which reads
// HOTWIRE_TURBO_STREAM_SECRET at process start and falls back to a
// freshly generated random key. That fallback works only for a single
// process — every replica would sign with its own key, so tokens issued
// by one node would fail verification on another, and any restart
// invalidates outstanding tokens. Horizontally scaled deployments must
// export HOTWIRE_TURBO_STREAM_SECRET with the same value on every
// process.
func StreamSSEHandler(sb StreamBroker, cfgs ...StreamConfig) http.Handler {
	c := &streamConfigs{
		heartbeatInterval: defaultHeartbeatInterval,
	}
	for _, cfg := range cfgs {
		cfg(c)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken := r.URL.Query().Get("token")
		if err := checkSameOrigin(r); err != nil {
			slog.Warn("turbo: sse authorize failed",
				"remote", r.RemoteAddr,
				"error", err,
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		stream, err := auth.VerifyToken(r, rawToken)
		if err != nil {
			slog.Warn("turbo: sse authorize failed",
				"remote", r.RemoteAddr,
				"error", err,
			)
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

func checkSameOrigin(r *http.Request) error {
	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		return errors.New("turbo: missing token query parameter")
	}

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

// writeEventStreamFormat writes one SSE event to w as an optional
// "event:" line, an optional "retry:" line, one "data:" line per newline
// segment of data, and a terminating blank line.
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

	for _, line := range bytes.Split(data, []byte("\n")) {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}
