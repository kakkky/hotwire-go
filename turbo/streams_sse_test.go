package turbo

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kakkky/hotwire-go/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSid is a canonical session identifier used across the tests to
// stand in for the value StreamSSEBroadcastMiddleware would install on a real
// request.
const testSid = "test-session-id"

// signedTestSid is testSid packaged as StreamSSEBroadcastMiddleware would put it
// on the wire: an auth.SignSid value that StreamSSEBroadcastHandler decodes
// with auth.VerifySid before comparing to the sid inside the token.
// Using a fresh long TTL keeps it valid across every test run.
var signedTestSid = auth.SignSid(testSid, defaultStreamTokenTTL)

func TestStreamSSEBroadcastSource(t *testing.T) {
	tests := []struct {
		name   string
		stream string
	}{
		{
			name:   "simple identifier",
			stream: "posts:42",
		},
		{
			name:   "chat room identifier",
			stream: "chat:room-abc",
		},
		{
			name:   "user notifications identifier",
			stream: "user:123:notifications",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(t.Context(), sidCtxKey{}, testSid)
			elm := StreamSSEBroadcastSource(ctx, tt.stream)

			assert.Equal(t, Tag("turbo-stream-source"), elm.Tag)
			require.Len(t, elm.Attrs, 1)

			attr := elm.Attrs[0]
			assert.Equal(t, "src", attr.Key)

			src, ok := attr.Value.(string)
			require.True(t, ok, "src attribute value must be a string")

			prefix := StreamSSEBroadcastPath + "?token="
			require.Truef(t, strings.HasPrefix(src, prefix), "src %q must start with %q", src, prefix)

			token := strings.TrimPrefix(src, prefix)
			payload, sid, err := auth.VerifyToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.stream, payload)
			assert.Equal(t, testSid, sid)
		})
	}
}

func TestStreamSSESource(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		src   string
		extra []Attrs
		want  Elm
	}{
		{
			name: "id and src",
			id:   "answer-source",
			src:  "/answers/42/stream",
			want: Elm{
				Tag: Tag("turbo-stream-source"),
				Attrs: Attrs{
					{Key: "id", Value: "answer-source"},
					{Key: "src", Value: "/answers/42/stream"},
				},
			},
		},
		{
			name:  "id and src with extra attrs",
			id:    "answer-source",
			src:   "/answers/42/stream",
			extra: []Attrs{{{Key: "data-answer", Value: "42"}}},
			want: Elm{
				Tag: Tag("turbo-stream-source"),
				Attrs: Attrs{
					{Key: "id", Value: "answer-source"},
					{Key: "src", Value: "/answers/42/stream"},
					{Key: "data-answer", Value: "42"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StreamSSESource(tt.id, tt.src, tt.extra...))
		})
	}
}

// nonFlushingResponseWriter hides the http.Flusher implementation of
// the embedded recorder.
type nonFlushingResponseWriter struct {
	http.ResponseWriter
}

// unwrappingResponseWriter hides the http.Flusher implementation of
// the embedded recorder but exposes the recorder through Unwrap.
type unwrappingResponseWriter struct {
	http.ResponseWriter
}

func (w unwrappingResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func TestNewStreamSSEWriter(t *testing.T) {
	sseHeader := http.Header{
		"Content-Type":      {"text/event-stream"},
		"Cache-Control":     {"no-cache"},
		"Connection":        {"keep-alive"},
		"X-Accel-Buffering": {"no"},
	}
	tests := []struct {
		name        string
		wrap        func(rec *httptest.ResponseRecorder) http.ResponseWriter
		wantErr     error
		wantHeader  http.Header
		wantFlushed bool
	}{
		{
			name:        "flushable writer gets event-stream headers",
			wrap:        func(rec *httptest.ResponseRecorder) http.ResponseWriter { return rec },
			wantHeader:  sseHeader,
			wantFlushed: true,
		},
		{
			name:        "wrapped writer is flushed through Unwrap",
			wrap:        func(rec *httptest.ResponseRecorder) http.ResponseWriter { return unwrappingResponseWriter{rec} },
			wantHeader:  sseHeader,
			wantFlushed: true,
		},
		{
			name:        "non-flushable writer is rejected untouched",
			wrap:        func(rec *httptest.ResponseRecorder) http.ResponseWriter { return nonFlushingResponseWriter{rec} },
			wantErr:     ErrStreamSSEFlushNotSupported,
			wantHeader:  http.Header{},
			wantFlushed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			_, err := NewStreamSSEWriter(tt.wrap(rec))

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantHeader, rec.Header())
			assert.Equal(t, tt.wantFlushed, rec.Flushed)
		})
	}
}

func TestStreamSSEWriter_Write(t *testing.T) {
	renderErr := errors.New("render failed")
	tests := []struct {
		name     string
		contents []StreamContent
		wantErr  error
		wantBody string
	}{
		{
			name:     "single content becomes one message event",
			contents: []StreamContent{StreamRemove("source")},
			wantBody: "event: message\ndata: <turbo-stream action=\"remove\" target=\"source\"><template></template></turbo-stream>\n\n",
		},
		{
			name: "multiple contents are concatenated into one event",
			contents: []StreamContent{
				encodedStreamContent("<a>"),
				encodedStreamContent("<b>"),
			},
			wantBody: "event: message\ndata: <a><b>\n\n",
		},
		{
			name:     "bare CR does not inject a new SSE event",
			contents: []StreamContent{encodedStreamContent("foo\rdata: injected\r\r")},
			wantBody: "event: message\ndata: foo\ndata: data: injected\ndata: \ndata: \n\n",
		},
		{
			name: "render error writes nothing",
			contents: []StreamContent{
				encodedStreamContent("<a>"),
				stubContent{err: renderErr},
			},
			wantErr:  renderErr,
			wantBody: "",
		},
		{
			name:     "no contents writes nothing",
			contents: nil,
			wantBody: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			sse, err := NewStreamSSEWriter(rec)
			require.NoError(t, err)

			err = sse.Write(t.Context(), tt.contents...)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestStreamSSEWriter_Heartbeat(t *testing.T) {
	rec := httptest.NewRecorder()
	sse, err := NewStreamSSEWriter(rec)
	require.NoError(t, err)

	require.NoError(t, sse.Heartbeat())

	assert.Equal(t, ":\n\n", rec.Body.String())
}

func TestStreamSSEBroadcastHandler(t *testing.T) {
	tests := []struct {
		name     string
		stream   string
		payload  []byte
		wantData string
	}{
		{
			name:     "single line turbo-stream fragment",
			stream:   "posts:42",
			payload:  []byte(`<turbo-stream action="append" target="messages"></turbo-stream>`),
			wantData: `<turbo-stream action="append" target="messages"></turbo-stream>`,
		},
		{
			name:     "multi-line payload joined by newlines",
			stream:   "chat:room-1",
			payload:  []byte("line1\nline2\nline3"),
			wantData: "line1\nline2\nline3",
		},
		{
			// Bare "\r" ends an SSE field per the HTML spec, so it must
			// be normalized to "\n" before the payload is written out.
			// Otherwise a "\r" from user content would let subsequent
			// bytes be interpreted as new SSE fields.
			name:     "bare CR is normalized to newline separator",
			stream:   "chat:room-2",
			payload:  []byte("line1\rline2"),
			wantData: "line1\nline2",
		},
		{
			name:     "CRLF is normalized to a single newline separator",
			stream:   "chat:room-3",
			payload:  []byte("line1\r\nline2"),
			wantData: "line1\nline2",
		},
		{
			// A crafted payload aimed at injecting a fresh SSE event via
			// bare "\r" must land inside one event's data-only lines.
			// Without normalization the browser would parse "data: injected"
			// as a new field and "\r\r" as an event terminator.
			name:     "bare CR does not inject a new SSE event",
			stream:   "chat:room-4",
			payload:  []byte("foo\rdata: injected\r\r"),
			wantData: "foo\ndata: injected\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			server := httptest.NewServer(StreamSSEBroadcastHandler(sb, WithStreamSSEBroadcastHeartbeatInterval(1*time.Hour)))
			defer server.Close()

			token := auth.SignToken(tt.stream, testSid, defaultStreamTokenTTL)
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"?token="+token, nil)
			require.NoError(t, err)
			req.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})

			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")

			// The handler subscribes before flushing response headers, so
			// once Do returns the subscription is already registered on
			// the broker and Publish reaches this connection deterministically.
			require.NoError(t, sb.Publish(context.Background(), tt.stream, tt.payload))

			// Read one SSE event: collect data: lines until a blank line.
			scanner := bufio.NewScanner(resp.Body)
			var dataLines []string
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" && len(dataLines) > 0 {
					break
				}
				if data, ok := strings.CutPrefix(line, "data: "); ok {
					dataLines = append(dataLines, data)
				}
			}
			require.NoError(t, scanner.Err())
			assert.Equal(t, tt.wantData, strings.Join(dataLines, "\n"))
		})
	}
}

// TestStreamSSEBroadcastHandler_Authorize covers the authorize step: malformed,
// missing, cookie-less, or sid-mismatched requests fail with 401; a
// valid cookie whose sid matches the sid baked into the token is
// accepted with 200 on both HTTP and HTTPS (r.TLS != nil branch).
func TestStreamSSEBroadcastHandler_Authorize(t *testing.T) {
	validSigned := auth.SignToken("posts:42", testSid, defaultStreamTokenTTL)

	tests := []struct {
		name       string
		useHTTPS   bool
		req        func(t *testing.T, serverURL string) *http.Request
		wantStatus int
	}{
		{
			name: "valid request over HTTP is accepted",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "valid request over HTTPS is accepted",
			useHTTPS: true,
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "missing token returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (no dot separator) returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token=not-a-real-token", nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (invalid base64 signed part) returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				token := "!!!not-base64!!!." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+token, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (invalid base64 signature) returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				token := base64.RawURLEncoding.EncodeToString([]byte("1\nsid\npayload")) +
					".!!!not-base64!!!"
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+token, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (signed missing exp newline) returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				token := base64.RawURLEncoding.EncodeToString([]byte("no-newlines-here")) +
					"." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+token, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (signed missing sid newline) returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				token := base64.RawURLEncoding.EncodeToString([]byte("1\nsid-only")) +
					"." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+token, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (non-numeric expiry) returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				token := base64.RawURLEncoding.EncodeToString([]byte("not-a-number\nsid\npayload")) +
					"." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+token, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			// Appending extra base64url chars extends the decoded
			// signature so it never matches the 32-byte HMAC output.
			name: "tampered signature returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned+"AA", nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired token returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				token := auth.SignToken("posts:42", testSid, -time.Hour)
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+token, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing session cookie returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned, nil)
				require.NoError(t, err)
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "cookie sid does not match token sid returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{
					Name:  streamsSessionCookieName,
					Value: auth.SignSid("different-session-id", defaultStreamTokenTTL),
				})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "cookie value that fails signature verification returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: "not-a-signed-sid"})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "cookie with tampered signature returns 401",
			req: func(t *testing.T, serverURL string) *http.Request {
				r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"?token="+validSigned, nil)
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid + "AA"})
				return r
			},
			wantStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			handler := StreamSSEBroadcastHandler(sb, WithStreamSSEBroadcastHeartbeatInterval(1*time.Hour))

			var server *httptest.Server
			if tt.useHTTPS {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			defer server.Close()

			resp, err := server.Client().Do(tt.req(t, server.URL))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// TestStreamSSEBroadcastHandler_Heartbeat drives the heartbeat.C branch of the
// handler's select loop deterministically. Inside a testing/synctest
// bubble the request context's fake-time deadline fires exactly `ticks`
// heartbeats before the handler exits via r.Context().Done().
func TestStreamSSEBroadcastHandler_Heartbeat(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		ticks    int
	}{
		{
			name:     "one interval fires one heartbeat comment",
			interval: time.Second,
			ticks:    1,
		},
		{
			name:     "three intervals fire three heartbeat comments",
			interval: time.Second,
			ticks:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				sb := NewStreamBroker()

				// Timeout mid-way between the Nth and (N+1)th tick so
				// exactly `ticks` heartbeats fire before ctx.Done exits
				// the handler.
				timeout := tt.interval*time.Duration(tt.ticks) + tt.interval/2
				ctx, cancel := context.WithTimeout(t.Context(), timeout)
				defer cancel()

				token := auth.SignToken("posts:42", testSid, defaultStreamTokenTTL)
				req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/turbo_streams_sse?token="+token, nil)
				req.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: signedTestSid})
				w := httptest.NewRecorder()

				StreamSSEBroadcastHandler(sb, WithStreamSSEBroadcastHeartbeatInterval(tt.interval)).ServeHTTP(w, req)

				assert.Equal(t, tt.ticks, strings.Count(w.Body.String(), ":\n\n"))
			})
		})
	}
}
