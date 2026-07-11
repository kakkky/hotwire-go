package turbo

import (
	"bufio"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			elm := StreamSourceSSE(tt.stream)

			assert.Equal(t, Tag("turbo-stream-source"), elm.Tag)
			require.Len(t, elm.Attrs, 1)

			attr := elm.Attrs[0]
			assert.Equal(t, "src", attr.Key)

			src, ok := attr.Value.(string)
			require.True(t, ok, "src attribute value must be a string")

			prefix := StreamsSSEPath + "?token="
			require.Truef(t, strings.HasPrefix(src, prefix), "src %q must start with %q", src, prefix)

			token := strings.TrimPrefix(src, prefix)
			decoded, err := verifyStreamToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.stream, decoded)
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

			token := signStreamToken(tt.stream, defaultStreamTokenTTL)
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"?token="+token, nil)
			require.NoError(t, err)

			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

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
	validSigned := signStreamToken("posts:42", defaultStreamTokenTTL)

	tests := []struct {
		name              string
		useHTTPS          bool
		token             string
		origin            string
		useMatchingOrigin bool
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
			name:       "tampered signature returns 401",
			token:      validSigned[:len(validSigned)-1] + "A",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token returns 401",
			token:      signStreamToken("posts:42", -time.Hour),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "cross-origin request returns 401",
			token:      signStreamToken("posts:42", defaultStreamTokenTTL),
			origin:     "https://evil.example.com",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:              "matching HTTPS Origin is accepted",
			useHTTPS:          true,
			token:             signStreamToken("posts:42", defaultStreamTokenTTL),
			useMatchingOrigin: true,
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

			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

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

				token := signStreamToken("posts:42", defaultStreamTokenTTL)
				req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/turbo_streams_sse?token="+token, nil)
				w := httptest.NewRecorder()

				StreamSSEHandler(sb, WithHeartbeatInterval(tt.interval)).ServeHTTP(w, req)

				assert.Equal(t, tt.ticks, strings.Count(w.Body.String(), ":\n\n"))
			})
		})
	}
}
