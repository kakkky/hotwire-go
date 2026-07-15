package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// hotwireGoSecretEnv names the environment variable that supplies
// the HMAC key used to sign and verify tokens.
const hotwireGoSecretEnv = "HOTWIRE_GO_SECRET"

// hotwireGoSecret is the process-wide HMAC key used by
// SignToken and VerifyToken.
//
// It is populated by init: HOTWIRE_GO_SECRET when set, otherwise
// a freshly generated 32-byte random key. The random fallback is only
// viable for a single-process deployment — every replica would sign with
// its own key, so tokens minted on one node fail verification on another,
// and every restart invalidates every outstanding token. Horizontally
// scaled deployments must export HOTWIRE_GO_SECRET with the same
// value on every process.
var hotwireGoSecret []byte

func init() {
	hotwireGoSecret = loadSecret(os.Getenv(hotwireGoSecretEnv))
}

// loadSecret returns the HMAC key material for tokens. A non-empty
// fromEnv is used verbatim; otherwise a freshly generated 32-byte
// random key is returned.
func loadSecret(fromEnv string) []byte {
	if fromEnv != "" {
		return []byte(fromEnv)
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("auth: failed to generate secret: " + err.Error())
	}
	return b
}

// SignToken mints a token that binds the given payload for ttl. The
// payload and an absolute expiry are HMAC-SHA256 signed with
// hotwireGoSecret and encoded as a URL-safe string suitable for a
// query parameter.
func SignToken(payload string, ttl time.Duration) string {
	token := token{
		payload:   payload,
		expiresAt: time.Now().Add(ttl).Unix(),
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(token.signingBytes())
	token.signature = h.Sum(nil)

	return token.encode()
}

// VerifyToken decodes the encoded form, recomputes the HMAC over the
// original payload, and rejects the token if the signature disagrees or the
// expiry has passed. The signature comparison uses hmac.Equal to keep it
// constant-time and immune to timing side channels.
func VerifyToken(token string) (string, error) {
	st, err := decodetoken(token)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(st.signingBytes())
	if !hmac.Equal(st.signature, h.Sum(nil)) {
		return "", errors.New("auth: token signature mismatch")
	}
	if time.Now().Unix() > st.expiresAt {
		return "", errors.New("auth: expired token")
	}
	return st.payload, nil
}

// token is the in-memory shape of a signed token: the payload, an
// absolute expiry in Unix seconds, and the HMAC-SHA256 signature
// (32 bytes) over the first two fields.
type token struct {
	payload   string
	expiresAt int64  // Unix seconds
	signature []byte // 32 bytes (HMAC-SHA256 output)
}

// signingBytes returns the exact bytes fed into HMAC —
// "payload\nexpiresAt" — so signing and verification always hash the
// identical input.
func (t token) signingBytes() []byte {
	return []byte(t.payload + "\n" + strconv.FormatInt(t.expiresAt, 10))
}

// encode serializes a signed token as "base64url(payload).base64url(sig)",
// URL-safe so it can ride in a query string without escaping.
func (t token) encode() string {
	return base64.RawURLEncoding.EncodeToString(t.signingBytes()) +
		"." +
		base64.RawURLEncoding.EncodeToString(t.signature)
}

// decodetoken parses the "base64url(payload).base64url(sig)" wire
// format back into a token. Any structural problem — missing
// separator, invalid base64, missing newline inside the payload, or
// non-numeric expiry — is collapsed into a single generic error so callers
// cannot use the failure mode as an oracle.
func decodetoken(s string) (token, error) {
	payloadPart, sigPart, ok := strings.Cut(s, ".")
	if !ok {
		return token{}, errors.New("auth: malformed token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return token{}, errors.New("auth: malformed token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return token{}, errors.New("auth: malformed token")
	}
	payloadBytes, expBytes, ok := bytes.Cut(payload, []byte{'\n'})
	if !ok {
		return token{}, errors.New("auth: malformed token")
	}
	exp, err := strconv.ParseInt(string(expBytes), 10, 64)
	if err != nil {
		return token{}, errors.New("auth: malformed token")
	}
	return token{
		payload:   string(payloadBytes),
		expiresAt: exp,
		signature: sig,
	}, nil
}
