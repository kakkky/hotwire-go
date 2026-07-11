package broker

import "github.com/redis/go-redis/v9"

type Config func(*configs)

type configs struct {
	redisClient redis.UniversalClient
}

func WithRedis(client redis.UniversalClient) Config {
	return func(c *configs) {
		c.redisClient = client
	}
}
