package turbo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

// streamsSessionCookieName is the cookie the middleware issues to
// carry the per-browser Turbo Streams session identifier (sid).
const streamsSessionCookieName = "turbo-streams-session"

// sidCtxKey types the ctx key used to expose the sid to downstream
// handlers and render-side helpers. Kept unexported so external
// packages cannot collide on the key or read the value out-of-band.
type sidCtxKey struct{}

// StreamsMiddleware issues a per-browser Turbo Streams session cookie
// and threads its value (sid) into r.Context. StreamSourceSSE bakes
// that sid into every token it mints, and StreamSSEHandler compares
// it against the cookie the browser presents when subscribing, so a
// token leaked to another browser cannot be replayed.
//
// Wrap the routes that render pages containing StreamSourceSSE. The
// SSE endpoint reads the cookie directly and does not need this
// middleware:
//
//	mux.Handle(turbo.StreamsSSEPath, turbo.StreamSSEHandler(sb))
//	mux.Handle("/", turbo.StreamsMiddleware(pageHandler))
func StreamsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sid string
		if cookie, err := r.Cookie(streamsSessionCookieName); err == nil && cookie.Value != "" {
			sid = cookie.Value
		} else {
			sid = genSid()
			http.SetCookie(w, &http.Cookie{
				Name:     streamsSessionCookieName,
				Value:    sid,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   r.TLS != nil,
			})
		}

		ctx := context.WithValue(r.Context(), sidCtxKey{}, sid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// genSid returns a fresh, URL-safe, 32-byte random session identifier
// for use as the streamsSessionCookieName value. crypto/rand backs the
// randomness so the identifier is unpredictable to attackers who do
// not observe the cookie directly.
func genSid() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("turbo: failed to generate stream secret: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
