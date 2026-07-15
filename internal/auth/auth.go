package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// hotwireTurboStreamSecretEnv names the environment variable that supplies
// the HMAC key used to sign and verify Turbo Streams subscription tokens.
const hotwireGoSecretEnv = "HOTWIRE_GO_SECRET"

// SessionCookieName is the cookie StreamsMiddleware issues to bind a
// Turbo Streams subscription token to a specific browser session.
// signStreamTokenWithSid folds HMAC(secret, cookieValue) into the signed
// payload; authorizeStreamRequest re-derives the same value from the
// cookie the browser sends on the SSE subscription request and rejects
// the request when it does not match.
const SessionCookieName = "_hotwire_go_sid"

// streamTokenVersion marks the leading claim of the signed payload.
// The sid claim slot was introduced with v1; keeping a version marker
// lets a future format change reject earlier tokens without ambiguity.
const tokenVersion = "v1"

// hotwireTurboStreamSecret is the process-wide HMAC key used by
// signStreamToken and verifyStreamToken.
//
// It is populated by init: HOTWIRE_TURBO_STREAM_SECRET when set, otherwise
// a freshly generated 32-byte random key. The random fallback is only
// viable for a single-process deployment — every replica would sign with
// its own key, so tokens minted on one node fail verification on another,
// and every restart invalidates every outstanding token. Horizontally
// scaled deployments must export HOTWIRE_TURBO_STREAM_SECRET with the same
// value on every process.
var hotwireGoSecret []byte

func init() {
	hotwireGoSecret = loadStreamSecret(os.Getenv(hotwireGoSecretEnv))
}

// loadStreamSecret returns the HMAC key material for stream tokens.
// A non-empty fromEnv is used verbatim; otherwise a freshly generated
// 32-byte random key is returned.
func loadStreamSecret(fromEnv string) []byte {
	if fromEnv != "" {
		return []byte(fromEnv)
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("turbo: failed to generate stream secret: " + err.Error())
	}
	return b
}

func SignToken(ctx context.Context, stream string, ttl time.Duration) string {
	sid, ok := ctx.Value(SidContextKey{}).(string)
	if !ok {
		return ""
	}
	token := streamToken{
		version:   tokenVersion,
		payload:   stream,
		expiresAt: time.Now().Add(ttl).Unix(),
		sid:       hmacSid(sid),
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(token.signingBytes())
	token.signature = h.Sum(nil)

	return token.encode()
}

func VerifyToken(r *http.Request, token string) (stream string, err error) {
	st, err := decodeToken(token)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(st.signingBytes())
	if !hmac.Equal(st.signature, h.Sum(nil)) {
		return "", errors.New("turbo: stream token signature mismatch")
	}

	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}
	// Compare the token sid to hmacSid(cookie.Value) rather than to the
	// raw cookie value: the token carries HMAC(secret, sid), so an
	// attacker who scrapes the URL cannot forge a cookie without also
	// knowing the raw sid. hmac.Equal keeps the comparison constant-time.
	if !hmac.Equal([]byte(st.sid), []byte(hmacSid(cookie.Value))) {
		return "", errors.New("turbo: stream token sid mismatch")
	}

	if time.Now().Unix() > st.expiresAt {
		return "", errors.New("turbo: expired stream token")
	}
	return st.payload, nil
}

// streamToken is the in-memory shape of a Turbo Streams subscription
// token. The wire form is "base64url(signingBytes).base64url(signature)",
// where signingBytes concatenates the four claim fields — version,
// payload, expiresAt, sid — separated by newlines and the signature is
// HMAC-SHA256 (32 bytes) over those same bytes.
type streamToken struct {
	version   string // format marker; must equal streamTokenVersion
	payload   string // stream name the token authorizes subscription to
	expiresAt int64  // absolute expiry in Unix seconds
	sid       string // hex(HMAC(secret, cookie sid))
	signature []byte // HMAC-SHA256 over signingBytes
}

// signingBytes returns the exact bytes fed into HMAC —
// "version\npayload\nexpiresAt\nsid" — so signing and verification
// always hash the identical input.
func (t streamToken) signingBytes() []byte {
	return []byte(
		t.version + "\n" +
			t.payload + "\n" +
			strconv.FormatInt(t.expiresAt, 10) + "\n" +
			t.sid,
	)
}

// encode serializes a signed token as "base64url(bytes).base64url(sig)",
// URL-safe so it can ride in a query string without escaping.
func (t streamToken) encode() string {
	return base64.RawURLEncoding.EncodeToString(t.signingBytes()) +
		"." +
		base64.RawURLEncoding.EncodeToString(t.signature)
}

// decodeStreamToken parses the "base64url(bytes).base64url(sig)" wire
// format back into a streamToken. Any structural problem — missing
// separator, invalid base64, wrong version marker, missing field, or
// non-numeric expiry — is collapsed into a single generic error so
// callers cannot use the failure mode as an oracle.
func decodeToken(s string) (streamToken, error) {
	bytesPart, sigPart, ok := strings.Cut(s, ".")
	if !ok {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(bytesPart)
	if err != nil {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	parts := bytes.SplitN(raw, []byte{'\n'}, 4)
	if len(parts) != 4 {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	if string(parts[0]) != tokenVersion {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	exp, err := strconv.ParseInt(string(parts[2]), 10, 64)
	if err != nil {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	return streamToken{
		version:   string(parts[0]),
		payload:   string(parts[1]),
		expiresAt: exp,
		sid:       string(parts[3]),
		signature: sig,
	}, nil
}

// hmacSid returns hex(HMAC-SHA256(secret, sid)) — the value stored in
// the token's sid claim. The raw sid never appears in the token itself,
// so an attacker who intercepts the URL sees only the HMAC output and
// cannot recover the cookie value needed to pass verifyStreamToken.
func hmacSid(sid string) string {
	if sid == "" {
		return ""
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write([]byte(sid))
	return hex.EncodeToString(h.Sum(nil))
}
