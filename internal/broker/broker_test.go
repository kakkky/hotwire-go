package broker

import (
	"testing"

	"github.com/redis/go-redis/v9"
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
		{
			name: "WithRedis returns a redis-backed broker",
			cfgs: []Config{WithRedis(redis.NewClient(&redis.Options{Addr: "localhost:0"}))},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.cfgs...)
			assert.NotNil(t, b)
		})
	}
}
