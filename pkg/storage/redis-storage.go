package storage

import (
	"time"

	"github.com/go-redis/redis"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	return &RedisStorage{
		client: redis.NewClient(&redis.Options{
			Addr:	  addr,
			Password: password,
			DB:		  db,
		}),
	}, nil
}

func (r *RedisStorage) Set(key string, value string, timeout time.Duration) error {
	return r.client.Set(key, value, timeout).Err()
}

func (r *RedisStorage) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *RedisStorage) Incr(key string) error {
	return r.client.Incr(key).Err()
}