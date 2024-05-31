package limiter

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/magalhaesgustavo/rate-limiter/cmd/configs"
	"github.com/magalhaesgustavo/rate-limiter/pkg/storage"

	"github.com/go-redis/redis"
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

func (limiter *RateLimiter) LimitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("API_KEY")

		ip := strings.Split(r.RemoteAddr, ":")[0]

		key, keyType := limiter.getKeyType(apiKey, ip)

		if !limiter.processRateLimiter(w, key, keyType) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (limiter *RateLimiter) processRateLimiter(w http.ResponseWriter, key, keyType string) bool {
	if limiter.isBlocked(key) {
		http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		return false
	}
	if limiter.checkRateLimit(key, keyType) {
		limiter.block(key, keyType)
		http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		return false
	}
	return true
}

func (limiter *RateLimiter) checkRateLimit(key string, keyType string) bool {
	val, err := limiter.redisClient.Get(key)

	if err == redis.Nil {
		log.Println("Key not found, starting counter: 1")
		if keyType == "ip" {
			limiter.redisClient.Set(key, "1", time.Duration(limiter.config.RequestsByIp)*time.Second)
		} else {
			limiter.redisClient.Set(key, "1", time.Duration(limiter.config.RequestsByToken)*time.Second)
		}
		limiter.redisClient.Incr(key)
		return false
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting counter value: %v\n", err)
		return false
	}

	log.Printf("Current counter: %d\n", count)

	reqLimit, blockTime := getLimitAndBlockTimePerType(keyType, limiter)

	if count > reqLimit {
		log.Println("Request counter exceeded, you are blocked for ", blockTime, " seconds")
		return true
	}

	limiter.redisClient.Incr(key)
	return false
}

func (limiter *RateLimiter) block(key string, tokenOrIp string) {
	var timeBlocked int
	if tokenOrIp == "ip" {
		timeBlocked = limiter.config.TimeBlockedByIp
	} else {
		timeBlocked = limiter.config.TimeBlockedByToken
	}
	limiter.redisClient.Set(key+":blocked", "1", time.Duration(timeBlocked)*time.Second)
}

func (limiter *RateLimiter) isBlocked(key string) bool {
	_, err := limiter.redisClient.Get(key + ":blocked")
	return err == nil
}

func (limiter *RateLimiter) getKeyType(apiKey string, ip string) (string, string) {
	var key, keyType string
	if apiKey == limiter.config.AllowedToken {
		key = "token:" + apiKey
		keyType = "token"
		log.Println("request by ", key)
	} else {
		key = "ip:" + ip
		keyType = "ip"
		log.Println("request by ", key)
	}
	return key, keyType
}

func getLimitAndBlockTimePerType(keyType string, limiter *RateLimiter) (int, int) {
	if keyType == "ip" {
		log.Println("total limit of requests per IP: ", limiter.config.RequestsByIp)
		return limiter.config.RequestsByIp, limiter.config.TimeBlockedByIp
	} 

	log.Println("total limit of requests per token: ", limiter.config.RequestsByToken)
	return limiter.config.RequestsByToken, limiter.config.TimeBlockedByToken
}
