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

// SignToken mints a URL-safe token that binds payload and sid for
// ttl. The expiry, sid, and payload are HMAC-SHA256 signed with the
// process-wide HMAC key and encoded so the result can ride in a query
// parameter.
//
// sid is embedded in the signed material so callers can bind the
// token to an identifier they hold independently: VerifyToken returns
// the sid alongside the payload for the caller to compare.
func SignToken(payload string, sid string, ttl time.Duration) string {
	token := token{
		payload:   payload,
		sid:       sid,
		expiresAt: time.Now().Add(ttl).Unix(),
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(token.signingBytes())
	token.signature = h.Sum(nil)

	return token.encode()
}

// VerifyToken decodes the wire format, recomputes the HMAC over the
// same material SignToken hashed, and rejects the token if the
// signature disagrees or the expiry has passed. The signature
// comparison uses hmac.Equal for constant-time safety.
//
// The sid embedded at signing time is returned alongside the payload
// for the caller to compare against its own reference.
func VerifyToken(token string) (payload, sid string, err error) {
	st, err := decodetoken(token)
	if err != nil {
		return "", "", err
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(st.signingBytes())
	if !hmac.Equal(st.signature, h.Sum(nil)) {
		return "", "", errors.New("auth: token signature mismatch")
	}
	if time.Now().Unix() > st.expiresAt {
		return "", "", errors.New("auth: expired token")
	}
	return st.payload, st.sid, nil
}

// token is the in-memory shape of a signed token: the payload, a
// session identifier the token is bound to, an absolute expiry in Unix
// seconds, and the HMAC-SHA256 signature (32 bytes) over expiresAt,
// sid, and payload concatenated by signingBytes.
type token struct {
	payload   string
	sid       string
	expiresAt int64  // Unix seconds
	signature []byte // 32 bytes (HMAC-SHA256 output)
}

// signingBytes returns the exact bytes fed into HMAC —
// "{expiresAt}\n{sid}\n{payload}" — so signing and verification always
// hash the identical input.
//
// Field order is deliberate: expiresAt (digits only) and sid (base64url
// alphabet) cannot contain a '\n', so the first two Cut boundaries in
// decodetoken are unambiguous no matter what bytes the caller puts in
// payload. Placing the payload (which may legitimately contain '\n')
// last prevents a delimiter-confusion attack where an attacker who
// influences the payload could shift the decoded sid boundary to
// impersonate another session.
func (t token) signingBytes() []byte {
	return []byte(strconv.FormatInt(t.expiresAt, 10) + "\n" + t.sid + "\n" + t.payload)
}

// encode serializes a signed token as "base64url(payload).base64url(sig)",
// URL-safe so it can ride in a query string without escaping.
func (t token) encode() string {
	return base64.RawURLEncoding.EncodeToString(t.signingBytes()) +
		"." +
		base64.RawURLEncoding.EncodeToString(t.signature)
}

// decodetoken parses the "base64url(signed).base64url(sig)" wire
// format back into a token, where signed is the
// "{expiresAt}\n{sid}\n{payload}" material HMAC-signed by
// signingBytes. Any structural problem — missing separator, invalid
// base64, missing newline between expiresAt/sid/payload, or non-numeric
// expiry — is collapsed into a single generic error so callers cannot
// use the failure mode as an oracle.
func decodetoken(s string) (token, error) {
	signedPart, sigPart, ok := strings.Cut(s, ".")
	if !ok {
		return token{}, errors.New("auth: malformed token")
	}
	signed, err := base64.RawURLEncoding.DecodeString(signedPart)
	if err != nil {
		return token{}, errors.New("auth: malformed token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return token{}, errors.New("auth: malformed token")
	}
	expBytes, rest, ok := bytes.Cut(signed, []byte{'\n'})
	if !ok {
		return token{}, errors.New("auth: malformed token")
	}
	sidBytes, payloadBytes, ok := bytes.Cut(rest, []byte{'\n'})
	if !ok {
		return token{}, errors.New("auth: malformed token")
	}
	exp, err := strconv.ParseInt(string(expBytes), 10, 64)
	if err != nil {
		return token{}, errors.New("auth: malformed token")
	}
	return token{
		payload:   string(payloadBytes),
		sid:       string(sidBytes),
		expiresAt: exp,
		signature: sig,
	}, nil
}
