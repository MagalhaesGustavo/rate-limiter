package limiter

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/magalhaesgustavo/rate-limiter/cmd/configs"
	"github.com/magalhaesgustavo/rate-limiter/pkg/storage"
)

type RateLimiter struct {
	redisClient *storage.RedisStorage
	config      *configs.Conf
}

func NewRateLimiter() *RateLimiter {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	redisClient, err := storage.NewRedisStorage(config.RedisHost, "", 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return &RateLimiter{
		redisClient: redisClient,
		config:      config,
	}
}

func (limiter *RateLimiter) RateLimitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("API_KEY")
		ip := strings.Split(r.RemoteAddr, ":")[0]

		key, keyType := limiter.getKey(apiKey, ip)

		if !limiter.processRequest(w, key, keyType) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (limiter *RateLimiter) getKey(apiKey, ip string) (string, string) {
	if apiKey == limiter.config.AllowedToken {
		return "token:" + apiKey, "token"
	}
	return "ip:" + ip, "ip"
}

func (limiter *RateLimiter) processRequest(w http.ResponseWriter, key, keyType string) bool {
	if limiter.isBlocked(key) {
		http.Error(w, "You have reached the maximum number of requests allowed within a certain time frame", http.StatusTooManyRequests)
		return false
	}
	if limiter.checkRateLimit(key, keyType) {
		limiter.block(key, keyType)
		http.Error(w, "You have reached the maximum number of requests allowed within a certain time frame", http.StatusTooManyRequests)
		return false
	}
	return true
}

func (limiter *RateLimiter) checkRateLimit(key, keyType string) bool {
	val, err := limiter.redisClient.Get(key)
	if err == redis.Nil {
		limiter.setInitialRequestCount(key, keyType)
		return false
	} else if err != nil {
		log.Printf("Failed to get key from Redis: %v", err)
		return false
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Failed to convert count to integer: %v", err)
		return false
	}

	requests := limiter.getRequestLimit(keyType)
	if count >= requests {
		return true
	}

	limiter.redisClient.Incr(key)
	return false
}

func (limiter *RateLimiter) setInitialRequestCount(key, keyType string) {
	duration := limiter.getRequestDuration(keyType)
	limiter.redisClient.Set(key, "1", duration)
}

func (limiter *RateLimiter) getRequestLimit(keyType string) int {
	if keyType == "ip" {
		return limiter.config.RequestsByIp
	}
	return limiter.config.RequestsByToken
}

func (limiter *RateLimiter) getRequestDuration(keyType string) time.Duration {
	if keyType == "ip" {
		return time.Duration(limiter.config.RequestsByIp) * time.Second
	}
	return time.Duration(limiter.config.RequestsByToken) * time.Second
}

func (limiter *RateLimiter) block(key, keyType string) {
	duration := limiter.getBlockDuration(keyType)
	limiter.redisClient.Set(key+":blocked", "1", duration)
}

func (limiter *RateLimiter) getBlockDuration(keyType string) time.Duration {
	if keyType == "ip" {
		return time.Duration(limiter.config.TimeBlockedByIp) * time.Second
	}
	return time.Duration(limiter.config.TimeBlockedByToken) * time.Second
}

func (limiter *RateLimiter) isBlocked(key string) bool {
	_, err := limiter.redisClient.Get(key + ":blocked")
	return err == nil
}
