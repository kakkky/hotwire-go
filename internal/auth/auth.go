package auth

import (
	"crypto/rand"
	"os"
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
