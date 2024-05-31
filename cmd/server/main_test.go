package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/magalhaesgustavo/rate-limiter/pkg/limiter"
	"github.com/magalhaesgustavo/rate-limiter/cmd/configs"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	router := setupRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()

	conf, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	// Test with a valid API_KEY
	t.Run("Valid API_KEY", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			resp, err := doRequest(ts, "GET", ts.URL, "token")
			if err != nil {
				t.Fatalf("Failed to make GET request: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if i < conf.RequestsByToken {
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK")
				assert.Equal(t, "Hello World!", string(body), "Expected response body 'Hello World!'")
			} else {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Expected status 429 Too Many Requests")
				assert.Equal(t, "you have reached the maximum number of requests or actions allowed within a certain time frame\n", string(body), "Expected rate limit message")
			}
		}
	})

	// Test with an invalid API_KEY
	t.Run("Invalid API_KEY", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			resp, err := doRequest(ts, "GET", ts.URL, "error")
			if err != nil {
				t.Fatalf("Failed to make GET request: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if i < conf.RequestsByIp {
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK")
				assert.Equal(t, "Hello World!", string(body), "Expected response body 'Hello World!'")
			} else {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Expected status 429 Too Many Requests")
				assert.Equal(t, "you have reached the maximum number of requests or actions allowed within a certain time frame\n", string(body), "Expected rate limit message")
			}
		}
	})

	// Wait for the rate limit to reset
	t.Logf("Waiting %v seconds to reset rate limit...", conf.TimeBlockedByIp)
	time.Sleep(time.Duration(conf.TimeBlockedByIp) * time.Second)

	// Test without API_KEY (IP based rate limiting)
	t.Run("IP Based Rate Limiting", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			resp, err := doRequest(ts, "GET", ts.URL, "")
			if err != nil {
				t.Fatalf("Failed to make GET request: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if i < conf.RequestsByIp {
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK")
				assert.Equal(t, "Hello World!", string(body), "Expected response body 'Hello World!'")
			} else {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Expected status 429 Too Many Requests")
				assert.Equal(t, "you have reached the maximum number of requests or actions allowed within a certain time frame\n", string(body), "Expected rate limit message")
			}
		}
	})
}


func setupRouter() *chi.Mux {
	router := chi.NewRouter()
	rateLimiter := limiter.NewRateLimiter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(rateLimiter.LimitHandler)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	return router
}

func doRequest(ts *httptest.Server, method, url, apiKey string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Add("API_KEY", apiKey)
	}

	return http.DefaultClient.Do(req)
}
