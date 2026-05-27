package tenants

import (
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// redisClientPool maintains a map of named *redis.Client instances so that
// we don't open duplicate connections to the same Redis endpoint.
type redisClientPool struct {
	clients sync.Map
}

func newRedisClientPool() *redisClientPool {
	return &redisClientPool{}
}

// RedisConnectionConfig holds the credentials needed to open a Redis connection.
type RedisConnectionConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Get returns an existing client keyed by name, or creates a new one.
func (p *redisClientPool) Get(name string, cfg RedisConnectionConfig) (*redis.Client, error) {
	if v, ok := p.clients.Load(name); ok {
		return v.(*redis.Client), nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	p.clients.Store(name, client)
	return client, nil
}

// getByOptions is a convenience wrapper used by the legacy getRedisClient method.
func (p *redisClientPool) getByOptions(name string, opts redis.Options) (*redis.Client, error) {
	if v, ok := p.clients.Load(name); ok {
		return v.(*redis.Client), nil
	}
	client := redis.NewClient(&opts)
	if client == nil {
		return nil, fmt.Errorf("failed to create Redis client for %s", name)
	}
	p.clients.Store(name, client)
	return client, nil
}
