package turbo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// loadStreamSecret is tested directly because init runs at package load
// and cannot be re-driven from a public-API test.

func TestLoadStreamSecret_FromEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want []byte
	}{
		{
			name: "uses env value verbatim",
			env:  "super-secret",
			want: []byte("super-secret"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, loadStreamSecret(tt.env))
		})
	}
}

func TestLoadStreamSecret_RandomFallback(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{
			name:    "returns a fresh 32-byte random key on every call",
			wantLen: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := loadStreamSecret("")
			second := loadStreamSecret("")

			assert.Len(t, first, tt.wantLen)
			assert.Len(t, second, tt.wantLen)
			assert.NotEqual(t, first, second)
		})
	}
}
