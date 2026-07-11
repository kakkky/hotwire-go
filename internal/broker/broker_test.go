package broker

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfgs    []Config
		wantNil bool
	}{
		{
			name: "no config returns default (in-process) broker",
			cfgs: nil,
		},
		{
			name:    "WithRedis picks the redis case (currently unimplemented, broker is nil)",
			cfgs:    []Config{WithRedis(redis.NewClient(&redis.Options{Addr: "localhost:0"}))},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.cfgs...)
			if tt.wantNil {
				assert.Nil(t, b)
			} else {
				assert.NotNil(t, b)
			}
		})
	}
}

