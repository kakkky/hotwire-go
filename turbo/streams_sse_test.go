package turbo

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
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

// extractSSESrcToken pulls the "token" query parameter out of the src
// attribute of a rendered <turbo-stream-source> tag. StreamSourceSSE
// resolves its attrs lazily at Render time, so tests must render into
// a buffer rather than inspect Elm.Attrs directly.
func extractSSESrcToken(t *testing.T, html string) string {
	t.Helper()
	srcPrefix := `src="` + StreamsSSEPath + `?token=`
	i := strings.Index(html, srcPrefix)
	require.GreaterOrEqualf(t, i, 0, "src prefix not found in %q", html)
	rest := html[i+len(srcPrefix):]
	j := strings.Index(rest, `"`)
	require.GreaterOrEqual(t, j, 0)
	return rest[:j]
}

// signWithSid signs a stream token with the given sid injected into ctx.
// signStreamToken now fails closed when ctx carries no sid, so tests
// that want a valid token — bound or intentionally unbound — must
// exercise the ctx path explicitly.
func signWithSid(t *testing.T, sid, stream string, ttl time.Duration) string {
	t.Helper()
	ctx := context.WithValue(t.Context(), auth.SidContextKey{}, sid)
	return signStreamToken(ctx, stream, ttl)
}

// verifyWithCookie builds a minimal *http.Request carrying the given
// cookie sid so verifyStreamToken can be driven from a unit test
// without wiring an httptest server. An empty cookieSid omits the
// cookie entirely, which matches the "no session cookie sent" wire
// state the SSE handler would see.
func verifyWithCookie(t *testing.T, token, cookieSid string) (string, error) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if cookieSid != "" {
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: cookieSid})
	}
	return verifyStreamToken(r, token)
}

func TestStreamSourceSSE(t *testing.T) {
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
			e := StreamSourceSSE(tt.stream)
			assert.Equal(t, Tag("turbo-stream-source"), e.Tag)

			// Render under a ctx that carries a fixed sid and verify
			// under a matching cookie. verifyStreamToken now requires
			// the session cookie to be present, so the "no cookie"
			// path is no longer a valid happy case.
			const sid = "stream-source-sse-test-sid"
			ctx := context.WithValue(t.Context(), auth.SidContextKey{}, sid)

			var buf bytes.Buffer
			require.NoError(t, e.Render(ctx, &buf))

			token := extractSSESrcToken(t, buf.String())
			decoded, err := verifyWithCookie(t, token, sid)
			require.NoError(t, err)
			assert.Equal(t, tt.stream, decoded)
		})
	}
}

// TestStreamSourceSSE_SidBinding covers the ctx-driven signing path:
// when the render ctx carries a sid (as it does when StreamsMiddleware
// is in the request chain), the emitted token verifies only against a
// caller whose cookie hashes to the same sid claim.
func TestStreamSourceSSE_SidBinding(t *testing.T) {
	tests := []struct {
		name       string
		sid        string
		verifyWith string
		wantErr    bool
	}{
		{
			name:       "matching cookie verifies",
			sid:        "session-alpha",
			verifyWith: "session-alpha",
		},
		{
			name:       "different cookie is rejected",
			sid:        "session-alpha",
			verifyWith: "session-beta",
			wantErr:    true,
		},
		{
			name:       "missing cookie against bound token is rejected",
			sid:        "session-alpha",
			verifyWith: "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(t.Context(), auth.SidContextKey{}, tt.sid)

			var buf bytes.Buffer
			require.NoError(t, StreamSourceSSE("posts:42").Render(ctx, &buf))

			token := extractSSESrcToken(t, buf.String())
			_, err := verifyWithCookie(t, token, tt.verifyWith)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStreamSSEHandler(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			server := httptest.NewServer(StreamSSEHandler(sb, WithHeartbeatInterval(1*time.Hour)))
			defer server.Close()

			// Sign under an explicit sid so the browser side can carry a
			// matching cookie; use a fixed test value rather than relying
			// on middleware in the httptest server.
			const sid = "sse-handler-test-sid"
			token := signWithSid(t, sid, tt.stream, defaultStreamTokenTTL)
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"?token="+token, nil)
			require.NoError(t, err)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sid})

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

// TestStreamSSEHandler_Authorize covers the authorize step: malformed /
// missing / cross-origin tokens fail with 401, and a matching HTTPS
// Origin (r.TLS != nil branch) is accepted with 200.
func TestStreamSSEHandler_Authorize(t *testing.T) {
	// Every valid-looking token in this table is bound to authSid so
	// the success case can supply a matching cookie; failure cases
	// intentionally skip the cookie so they exercise the "no session
	// cookie present" branch of verifyStreamToken.
	const authSid = "authorize-test-sid"
	validSigned := signWithSid(t, authSid, "posts:42", defaultStreamTokenTTL)

	tests := []struct {
		name              string
		useHTTPS          bool
		token             string
		origin            string
		useMatchingOrigin bool
		cookieSid         string
		wantStatus        int
	}{
		{
			name:       "missing token returns 401",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed token (no dot separator) returns 401",
			token:      "not-a-real-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed token (invalid base64 payload) returns 401",
			token:      "!!!not-base64!!!." + base64.RawURLEncoding.EncodeToString([]byte("sig")),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (invalid base64 signature) returns 401",
			token: base64.RawURLEncoding.EncodeToString([]byte("stream\n1")) +
				".!!!not-base64!!!",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (payload missing newline) returns 401",
			token: base64.RawURLEncoding.EncodeToString([]byte("no-newline-here")) +
				"." + base64.RawURLEncoding.EncodeToString([]byte("sig")),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token (non-numeric expiry) returns 401",
			token: base64.RawURLEncoding.EncodeToString([]byte("stream\nnot-a-number")) +
				"." + base64.RawURLEncoding.EncodeToString([]byte("sig")),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "tampered signature returns 401",
			// Appending extra base64url chars extends the decoded
			// signature so it never matches the 32-byte HMAC output.
			token:      validSigned + "AA",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token returns 401",
			token:      signWithSid(t, authSid, "posts:42", -time.Hour),
			cookieSid:  authSid,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "cross-origin request returns 401",
			token:      validSigned,
			origin:     "https://evil.example.com",
			cookieSid:  authSid,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:              "matching HTTPS Origin is accepted",
			useHTTPS:          true,
			token:             validSigned,
			useMatchingOrigin: true,
			cookieSid:         authSid,
			wantStatus:        http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			handler := StreamSSEHandler(sb, WithHeartbeatInterval(1*time.Hour))

			var server *httptest.Server
			if tt.useHTTPS {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			defer server.Close()

			url := server.URL
			if tt.token != "" {
				url += "?token=" + tt.token
			}

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
			require.NoError(t, err)
			switch {
			case tt.useMatchingOrigin:
				req.Header.Set("Origin", server.URL)
			case tt.origin != "":
				req.Header.Set("Origin", tt.origin)
			}
			if tt.cookieSid != "" {
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tt.cookieSid})
			}

			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// TestStreamSSEHandler_Authorize_CookieBinding covers the sid-bound
// path: authorizeStreamRequest reads the _hotwire_sid cookie and
// requires the incoming token's sid claim to hash to the same value.
func TestStreamSSEHandler_Authorize_CookieBinding(t *testing.T) {
	tests := []struct {
		name       string
		signSid    string
		cookieSid  string
		wantStatus int
	}{
		{
			name:       "matching cookie is accepted",
			signSid:    "session-alpha",
			cookieSid:  "session-alpha",
			wantStatus: http.StatusOK,
		},
		{
			name:       "different cookie is rejected",
			signSid:    "session-alpha",
			cookieSid:  "session-beta",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing cookie against bound token is rejected",
			signSid:    "session-alpha",
			cookieSid:  "",
			wantStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStreamBroker()
			server := httptest.NewServer(StreamSSEHandler(sb, WithHeartbeatInterval(1*time.Hour)))
			defer server.Close()

			token := signWithSid(t, tt.signSid, "posts:42", defaultStreamTokenTTL)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"?token="+token, nil)
			require.NoError(t, err)
			if tt.cookieSid != "" {
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tt.cookieSid})
			}

			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// TestStreamSSEHandler_Heartbeat drives the heartbeat.C branch of the
// handler's select loop deterministically. Inside a testing/synctest
// bubble the request context's fake-time deadline fires exactly `ticks`
// heartbeats before the handler exits via r.Context().Done().
func TestStreamSSEHandler_Heartbeat(t *testing.T) {
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

				const sid = "heartbeat-test-sid"
				token := signWithSid(t, sid, "posts:42", defaultStreamTokenTTL)
				req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/turbo_streams_sse?token="+token, nil)
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sid})
				w := httptest.NewRecorder()

				StreamSSEHandler(sb, WithHeartbeatInterval(tt.interval)).ServeHTTP(w, req)

				assert.Equal(t, tt.ticks, strings.Count(w.Body.String(), ":\n\n"))
			})
		})
	}
}
