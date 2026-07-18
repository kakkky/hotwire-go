package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignSid_VerifySid(t *testing.T) {
	tests := []struct {
		name    string
		input   func() string
		wantSid string
		wantErr string
	}{
		{
			name: "roundtrip returns the signed sid",
			input: func() string {
				return SignSid("Ql59aE8Tfn61lI8xIDNu9b3g9cUwpnjXLz_A_gZqd40", time.Hour)
			},
			wantSid: "Ql59aE8Tfn61lI8xIDNu9b3g9cUwpnjXLz_A_gZqd40",
		},
		{
			// Splice a signed part from one sid onto the signature
			// of another: both halves stay valid base64url so decode
			// succeeds, but the HMAC no longer matches.
			name: "tampered signature fails HMAC equality",
			input: func() string {
				signed := SignSid("sid", time.Hour)
				other := SignSid("other", time.Hour)
				signedPart := strings.SplitN(signed, ".", 2)[0]
				otherSig := strings.SplitN(other, ".", 2)[1]
				return signedPart + "." + otherSig
			},
			wantErr: "auth: sid signature mismatch",
		},
		{
			name: "negative ttl yields an already-expired sid",
			input: func() string {
				return SignSid("sid", -time.Hour)
			},
			wantErr: "auth: expired sid",
		},
		{
			name:    "no dot separator",
			input:   func() string { return "not-a-signed-sid" },
			wantErr: "auth: malformed sid",
		},
		{
			name: "invalid base64url in signed part",
			input: func() string {
				return "!!!not-base64!!!." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
			},
			wantErr: "auth: malformed sid",
		},
		{
			name: "invalid base64url in signature part",
			input: func() string {
				return base64.RawURLEncoding.EncodeToString([]byte("1\nsid")) + ".!!!not-base64!!!"
			},
			wantErr: "auth: malformed sid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifySid(tt.input())
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSid, got)
		})
	}
}
