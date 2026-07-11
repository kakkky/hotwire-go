package broker

import "github.com/redis/go-redis/v9"

// Config customizes how New builds a Broker. It follows the functional
// options pattern: each Config mutates an internal configuration
// struct, and New applies them in order before selecting a backend.
// Values are produced by backend-specific factory helpers such as
// WithRedis rather than constructed directly.
type Config func(*configs)

type configs struct {
	redisClient redis.UniversalClient
}

// WithRedis returns a Config that routes New's returned Broker
// through Redis PUB/SUB instead of the in-process default. The client
// is used as-is — pass an already-configured *redis.Client (or any
// implementation of redis.UniversalClient) so its connection pool and
// options can be shared with the rest of the application. The Broker
// only calls Publish and Subscribe on it, so a single-node client, a
// ClusterClient, and a FailoverClient are all acceptable inputs.
func WithRedis(client redis.UniversalClient) Config {
	return func(c *configs) {
		c.redisClient = client
	}
}
