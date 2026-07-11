package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		cfgs []Config
	}{
		{
			name: "no config returns default (in-process) broker",
			cfgs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.cfgs...)
			assert.NotNil(t, b)
		})
	}
}
