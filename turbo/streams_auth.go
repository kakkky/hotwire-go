package turbo

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// hotwireTurboStreamSecretEnv names the environment variable that supplies
// the HMAC key used to sign and verify Turbo Streams subscription tokens.
const hotwireTurboStreamSecretEnv = "HOTWIRE_TURBO_STREAM_SECRET"

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
var hotwireTurboStreamSecret []byte

func init() {
	hotwireTurboStreamSecret = loadStreamSecret(os.Getenv(hotwireTurboStreamSecretEnv))
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

// streamToken is the in-memory shape of a Turbo Streams subscription
// token: the target stream, an absolute expiry in Unix seconds, and the
// HMAC-SHA256 signature (32 bytes) over the first two fields.
type streamToken struct {
	stream    string
	expiresAt int64  // Unix seconds
	signature []byte // 32 bytes (HMAC-SHA256 output)
}

// signedPayload returns the exact bytes fed into HMAC — "stream\nexpiresAt"
// — so signing and verification always hash the identical input.
func (t streamToken) signedPayload() []byte {
	return []byte(t.stream + "\n" + strconv.FormatInt(t.expiresAt, 10))
}

// encode serializes a signed token as "base64url(payload).base64url(sig)",
// URL-safe so it can ride in a query string without escaping.
func (t streamToken) encode() string {
	return base64.RawURLEncoding.EncodeToString(t.signedPayload()) +
		"." +
		base64.RawURLEncoding.EncodeToString(t.signature)
}

// authorizeStreamRequest extracts the "token" query parameter, verifies its
// signature and expiry, and — when the request carries an Origin header —
// rejects it unless Origin matches the request's own scheme://host. The
// returned stream is the value that was baked into the token at sign time,
// so downstream code can subscribe to it directly without trusting any
// client-controlled input.
func authorizeStreamRequest(r *http.Request) (stream string, err error) {
	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		return "", errors.New("turbo: missing stream token")
	}

	stream, err = verifyStreamToken(rawToken)
	if err != nil {
		return "", err
	}

	if origin := r.Header.Get("Origin"); origin != "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if origin != scheme+"://"+r.Host {
			return "", errors.New("turbo: cross-origin request rejected")
		}
	}

	return stream, nil
}

// signStreamToken mints a token that authorizes subscription to the given
// stream for ttl. The stream name and an absolute expiry are HMAC-SHA256
// signed with hotwireTurboStreamSecret and encoded as a URL-safe string
// suitable for a query parameter.
func signStreamToken(stream string, ttl time.Duration) string {
	token := streamToken{
		stream:    stream,
		expiresAt: time.Now().Add(ttl).Unix(),
	}
	h := hmac.New(sha256.New, hotwireTurboStreamSecret)
	h.Write(token.signedPayload())
	token.signature = h.Sum(nil)

	return token.encode()
}

// verifyStreamToken decodes the encoded form, recomputes the HMAC over the
// original payload, and rejects the token if the signature disagrees or the
// expiry has passed. The signature comparison uses hmac.Equal to keep it
// constant-time and immune to timing side channels.
func verifyStreamToken(token string) (stream string, err error) {
	st, err := decodeStreamToken(token)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, hotwireTurboStreamSecret)
	h.Write(st.signedPayload())
	if !hmac.Equal(st.signature, h.Sum(nil)) {
		return "", errors.New("turbo: stream token signature mismatch")
	}
	if time.Now().Unix() > st.expiresAt {
		return "", errors.New("turbo: expired stream token")
	}
	return st.stream, nil
}

// decodeStreamToken parses the "base64url(payload).base64url(sig)" wire
// format back into a streamToken. Any structural problem — missing
// separator, invalid base64, missing newline inside the payload, or
// non-numeric expiry — is collapsed into a single generic error so callers
// cannot use the failure mode as an oracle.
func decodeStreamToken(s string) (streamToken, error) {
	payloadPart, sigPart, ok := strings.Cut(s, ".")
	if !ok {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	streamBytes, expBytes, ok := bytes.Cut(payload, []byte{'\n'})
	if !ok {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	exp, err := strconv.ParseInt(string(expBytes), 10, 64)
	if err != nil {
		return streamToken{}, errors.New("turbo: malformed stream token")
	}
	return streamToken{
		stream:    string(streamBytes),
		expiresAt: exp,
		signature: sig,
	}, nil
}
