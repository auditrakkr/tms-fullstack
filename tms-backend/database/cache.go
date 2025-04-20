package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)



type RedisCache struct {
	client *redis.Client
	ctx context.Context
}


// NewRedisCache creates a new RedisCache instance
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
		ctx: context.Background(),
	}
}


func (c *RedisCache) Get(key string, target interface{}) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("cache client is not initialized")
	}

	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // Key does not exist
		}
		return false, err // Some other error occurred
	}

	// Unmarshal the value into the target variable
	err = json.Unmarshal([]byte(val), target)
	if err != nil {
		return false, err // JSON unmarshal error
	}
	return true, nil // Key exists and value is unmarshalled successfully
}


// Set stores a value in the cache with expiration time in milliseconds
func (c *RedisCache) Set(key string, value interface{}, expiration int64) error {
	if c.client == nil {
		return fmt.Errorf("cache client is not initialized")
	}

	// Marshal the value to JSON
	val, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	// Set the value in the cache with expiration
	err = c.client.Set(c.ctx, key, val, time.Duration(expiration)*time.Millisecond).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in cache: %v", err)
	}

	return nil // Value set successfully
}

// Delete removes a key from the cache
func (c *RedisCache) Delete(key string) error {
	if c.client == nil {
		return fmt.Errorf("cache client is not initialized")
	}

	// Delete the key from the cache
	err := c.client.Del(c.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from cache: %v", err)
	}

	return nil // Key deleted successfully
}

// DeletePattern removes all keys matching a pattern
func (c *RedisCache) DeletePattern(pattern string) error {
	if c.client == nil {
		return fmt.Errorf("cache client is not initialized")
	}

	// Get all keys matching the pattern
	keys, err := c.client.Keys(c.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %v", err)
	}

	// Delete each key
	if len(keys) > 0 {
		err = c.client.Pipeline().Del(c.ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys by pattern: %v", err)
		}
	}

	return nil // Keys deleted successfully
}