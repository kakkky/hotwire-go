package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignToken_VerifyToken(t *testing.T) {
	tests := []struct {
		name        string
		input       func() string
		wantPayload string
		wantSid     string
		wantErr     string
	}{
		{
			name: "roundtrip returns the signed payload and sid",
			input: func() string {
				return SignToken("posts:42", "Ql59aE8Tfn61lI8xIDNu9b3g9cUwpnjXLz_A_gZqd40", time.Hour)
			},
			wantPayload: "posts:42",
			wantSid:     "Ql59aE8Tfn61lI8xIDNu9b3g9cUwpnjXLz_A_gZqd40",
		},
		{
			// Splice a signed part from one token onto the signature
			// of another: both halves stay valid base64url so decode
			// succeeds, but the HMAC no longer matches.
			name: "tampered signature fails HMAC equality",
			input: func() string {
				signed := SignToken("posts:42", "sid", time.Hour)
				other := SignToken("posts:99", "sid", time.Hour)
				signedPart := strings.SplitN(signed, ".", 2)[0]
				otherSig := strings.SplitN(other, ".", 2)[1]
				return signedPart + "." + otherSig
			},
			wantErr: "auth: token signature mismatch",
		},
		{
			name: "negative ttl yields an already-expired token",
			input: func() string {
				return SignToken("posts:42", "sid", -time.Hour)
			},
			wantErr: "auth: expired token",
		},
		{
			name:    "no dot separator",
			input:   func() string { return "not-a-real-token" },
			wantErr: "auth: malformed token",
		},
		{
			name: "invalid base64url in signed part",
			input: func() string {
				return "!!!not-base64!!!." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
			},
			wantErr: "auth: malformed token",
		},
		{
			name: "invalid base64url in signature part",
			input: func() string {
				return base64.RawURLEncoding.EncodeToString([]byte("1\nsid\npayload")) + ".!!!not-base64!!!"
			},
			wantErr: "auth: malformed token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, sid, err := VerifyToken(tt.input())
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPayload, payload)
			assert.Equal(t, tt.wantSid, sid)
		})
	}
}
