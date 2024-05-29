package storage

import "time"

type Storage interface {
	Set(key string, value string, timeout time.Duration) error
	Get(key string) (string, error)
	Incr(key string) error
}