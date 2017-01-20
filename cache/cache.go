package cache

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Cache is the interface to interact with an underlying cache
type Cache interface {
	SetField(string, string, string) error
	DeleteField(string, string) error
	Get(string) ([]byte, error)
	GetAll(key string) (map[string]string, error)
	Subscribe(string, func()) error
}

// PoolWrapper contains a pointer to the underlying connection pool.
type PoolWrapper struct {
	pool *redis.Pool
}

// SetField sets a key to a value
func (c *PoolWrapper) SetField(key, field, value string) error {
	conn := c.pool.Get()
	defer conn.Close()

	conn.Send("PUBLISH", "go-ocelot", "updated")
	if _, err := conn.Do("HSET", key, field, value); err != nil {
		return err
	}
	return nil
}

// DeleteField removes a hash from a key
func (c *PoolWrapper) DeleteField(key, field string) error {
	conn := c.pool.Get()
	defer conn.Close()

	conn.Send("PUBLISH", "go-ocelot", "updated")
	if _, err := conn.Do("HDEL", key, field); err != nil {
		return err
	}
	return nil
}

// Subscribe subscribes to a key and calls a function when messages are received
func (c *PoolWrapper) Subscribe(channel string, action func()) error {
	conn := c.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("SUBSCRIBE", "go-ocelot"); err != nil {
		return err
	}

	for {
		if _, err := conn.Receive(); err != nil {
			return err
		}

		action()
	}
}

// Get retrieves a key from redis
func (c *PoolWrapper) Get(key string) ([]byte, error) {
	var s []byte

	conn := c.pool.Get()
	defer conn.Close()

	result, err := conn.Do("HGETALL", key)
	if err != nil {
		return s, err
	}

	b, ok := result.([]byte)
	if !ok {
		return s, errors.New("Failed to parse value as a string")
	}

	return b, nil
}

// GetAll hash fields from a key in redis
func (c *PoolWrapper) GetAll(key string) (map[string]string, error) {
	conn := c.pool.Get()
	defer conn.Close()

	result, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return result, err
	}

	return result, nil
}

// New returns an initialized instance of a cache.
func New(address string) Cache {
	pool := &redis.Pool{
		MaxIdle:     2,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return Cache(&PoolWrapper{pool: pool})
}
