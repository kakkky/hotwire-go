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

// SignSid mints a signed cookie value that binds sid to ttl.
func SignSid(sid string, ttl time.Duration) string {
	expiresAt := time.Now().Add(ttl).Unix()
	signed := []byte(strconv.FormatInt(expiresAt, 10) + "\n" + sid)
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(signed)
	return base64.RawURLEncoding.EncodeToString(signed) +
		"." +
		base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// VerifySid decodes a value produced by SignSid, recomputes the
// HMAC, and rejects tampered or expired inputs. Structural failure
// modes are collapsed into a single generic error so callers cannot
// distinguish them.
func VerifySid(signedSid string) (sid string, err error) {
	signedPart, sigPart, ok := strings.Cut(signedSid, ".")
	if !ok {
		return "", errors.New("auth: malformed sid")
	}
	signed, err := base64.RawURLEncoding.DecodeString(signedPart)
	if err != nil {
		return "", errors.New("auth: malformed sid")
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return "", errors.New("auth: malformed sid")
	}
	h := hmac.New(sha256.New, hotwireGoSecret)
	h.Write(signed)
	if !hmac.Equal(sig, h.Sum(nil)) {
		return "", errors.New("auth: sid signature mismatch")
	}
	expBytes, sidBytes, ok := bytes.Cut(signed, []byte{'\n'})
	if !ok {
		return "", errors.New("auth: malformed sid")
	}
	exp, err := strconv.ParseInt(string(expBytes), 10, 64)
	if err != nil {
		return "", errors.New("auth: malformed sid")
	}
	if time.Now().Unix() > exp {
		return "", errors.New("auth: expired sid")
	}
	return string(sidBytes), nil
}
