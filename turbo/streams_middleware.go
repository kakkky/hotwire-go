package turbo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/kakkky/hotwire-go/internal/auth"
)

// StreamsMiddleware issues (and, on subsequent requests, propagates) a
// per-browser session id used to bind Turbo Streams subscription
// tokens. Wrap the top-level HTTP handler with it once:
//
//	mux := http.NewServeMux()
//	mux.Handle(...)
//	server.Handler = turbo.StreamsMiddleware(mux)
//
// On the first request from a given browser, the middleware issues a
// fresh 32-byte random sid as an HttpOnly, SameSite=Strict cookie
// (Secure when TLS is in use). On subsequent requests it reads the
// existing cookie. Either way, the sid is placed on the request context
// so StreamSourceSSE folds hmacSid(sid) into the sid claim of the
// token it emits, and verifyStreamToken cross-checks that value against
// hmacSid(cookie.Value) on the SSE subscription request.
//
// Deployments that skip this middleware get empty tokens from
// signStreamToken (fail-closed): any SSE subscribe attempt reaches
// authorizeStreamRequest with a blank query parameter and is rejected
// with 401. Mount the middleware on every request path that renders a
// <turbo-stream-source>.
func StreamsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sid string
		// r.Cookie returns (nil, http.ErrNoCookie) when the cookie is
		// absent, so dereferencing cookie.Value without checking err
		// would panic on the first request from a given browser.
		if cookie, err := r.Cookie(auth.SessionCookieName); err == nil && cookie.Value != "" {
			sid = cookie.Value
		} else {
			sid = genSid()
			http.SetCookie(w, &http.Cookie{
				Name:     auth.SessionCookieName,
				Value:    sid,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   r.TLS != nil,
			})
		}

		ctx := context.WithValue(r.Context(), auth.SidContextKey{}, sid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func genSid() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("turbo: failed to generate stream secret: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
