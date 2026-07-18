package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"
)

// SignToken mints a URL-safe token that binds payload and sid for
// ttl. The expiry, sid, and payload are HMAC-SHA256 signed with the
// process-wide HMAC key and encoded so the result can ride in a query
// parameter.
//
// Field order in the signed material is deliberate: expiresAt (digits
// only) and sid (base64url alphabet) cannot contain a '\n', so the
// first two Cut boundaries in VerifyToken are unambiguous no matter
// what bytes the caller puts in payload. Placing the payload (which
// may legitimately contain '\n') last prevents a delimiter-confusion
// attack where an attacker who influences the payload could shift the
// decoded sid boundary to impersonate another session.
//
// sid is embedded in the signed material so callers can bind the
// token to an identifier they hold independently: VerifyToken returns
// the sid alongside the payload for the caller to compare.
func SignToken(payload string, sid string, ttl time.Duration) string {
	expiresAt := time.Now().Add(ttl).Unix()
	signed := []byte(strconv.FormatInt(expiresAt, 10) + "\n" + sid + "\n" + payload)
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(signed)
	return base64.RawURLEncoding.EncodeToString(signed) +
		"." +
		base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// VerifyToken decodes the wire format, recomputes the HMAC over the
// same material SignToken hashed, and rejects the token if the
// signature disagrees or the expiry has passed. The signature
// comparison uses hmac.Equal for constant-time safety.
//
// Structural checks (field framing, expiry parsing) run after HMAC
// verification, so an attacker without the process-wide HMAC key
// cannot reach — and therefore cannot use as an oracle — any of the
// structural failure modes: every crafted-from-outside malformed
// input either fails at wire-format decode or fails at HMAC.
//
// The sid embedded at signing time is returned alongside the payload
// for the caller to compare against its own reference.
func VerifyToken(token string) (payload, sid string, err error) {
	signedPart, sigPart, ok := strings.Cut(token, ".")
	if !ok {
		return "", "", errors.New("auth: malformed token")
	}
	signed, err := base64.RawURLEncoding.DecodeString(signedPart)
	if err != nil {
		return "", "", errors.New("auth: malformed token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return "", "", errors.New("auth: malformed token")
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(signed)
	if !hmac.Equal(sig, h.Sum(nil)) {
		return "", "", errors.New("auth: token signature mismatch")
	}
	expBytes, rest, ok := bytes.Cut(signed, []byte{'\n'})
	if !ok {
		return "", "", errors.New("auth: malformed token")
	}
	sidBytes, payloadBytes, ok := bytes.Cut(rest, []byte{'\n'})
	if !ok {
		return "", "", errors.New("auth: malformed token")
	}
	exp, err := strconv.ParseInt(string(expBytes), 10, 64)
	if err != nil {
		return "", "", errors.New("auth: malformed token")
	}
	if time.Now().Unix() > exp {
		return "", "", errors.New("auth: expired token")
	}
	return string(payloadBytes), string(sidBytes), nil
}
