package turbo

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kakkky/hotwire-go/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamsMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		req           func() *http.Request
		wantSetCookie bool
		wantSecure    bool
		wantSid       string // "" means freshly minted (assert non-empty); non-empty means the reused raw sid
	}{
		{
			name: "no cookie mints a new sid over HTTP (Secure=false)",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			wantSetCookie: true,
		},
		{
			name: "no cookie mints a new sid over HTTPS (Secure=true)",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.TLS = &tls.ConnectionState{}
				return r
			},
			wantSetCookie: true,
			wantSecure:    true,
		},
		{
			name: "cookie with valid signed value is reused without issuing a new cookie",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.AddCookie(&http.Cookie{
					Name:  streamsSessionCookieName,
					Value: auth.SignSid("reused-sid", streamsSessionCookieTTL),
				})
				return r
			},
			wantSid: "reused-sid",
		},
		{
			name: "cookie with tampered value is replaced with a fresh sid",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.AddCookie(&http.Cookie{Name: streamsSessionCookieName, Value: "not-a-signed-sid"})
				return r
			},
			wantSetCookie: true,
		},
		{
			name: "cookie with expired signed value is replaced with a fresh sid",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.AddCookie(&http.Cookie{
					Name:  streamsSessionCookieName,
					Value: auth.SignSid("expired-sid", -time.Hour),
				})
				return r
			},
			wantSetCookie: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotSid string
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotSid, _ = r.Context().Value(sidCtxKey{}).(string)
			})
			w := httptest.NewRecorder()

			StreamsMiddleware(next).ServeHTTP(w, tt.req())

			if tt.wantSid == "" {
				assert.NotEmpty(t, gotSid, "sid must be threaded into ctx")
			} else {
				assert.Equal(t, tt.wantSid, gotSid)
			}

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()
			cookies := resp.Cookies()

			if tt.wantSetCookie {
				require.Len(t, cookies, 1)
				c := cookies[0]
				assert.Equal(t, streamsSessionCookieName, c.Name)
				verifiedFromCookie, err := auth.VerifySid(c.Value)
				require.NoError(t, err, "cookie value must be a valid signed sid")
				assert.Equal(t, gotSid, verifiedFromCookie, "cookie must sign the ctx sid")
				assert.Equal(t, "/", c.Path)
				assert.True(t, c.HttpOnly)
				assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
				assert.Equal(t, tt.wantSecure, c.Secure)
			} else {
				assert.Empty(t, cookies, "existing cookie must not trigger a Set-Cookie")
			}
		})
	}
}
