// Package turbo provides server-side helpers for building web
// applications that use Hotwire's Turbo runtime (Drive, Frames, Streams)
// with Go and html/template.
//
// The package is a thin server-side layer: it does not ship any of the
// Turbo JavaScript itself, and does not attempt to replace an HTTP mux
// or web framework.
//
// # Turbo Streams over SSE — production note
//
// Subscription tokens minted by StreamSourceSSE and verified by
// StreamSSEHandler are signed with an HMAC key read from
// HOTWIRE_GO_SECRET. When unset, a random per-process key is used —
// fine for a single-process development setup, but every replica in
// a horizontally scaled deployment must export the same
// HOTWIRE_GO_SECRET value; otherwise a token minted on one node
// fails verification on another.
package turbo
