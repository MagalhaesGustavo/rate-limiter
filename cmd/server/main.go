package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/magalhaesgustavo/rate-limiter/cmd/configs"
	"github.com/magalhaesgustavo/rate-limiter/pkg/limiter"
)

func main() {
	conf, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()
	rateLimiter := limiter.NewRateLimiter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(rateLimiter.RateLimitHandler)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	
	log.Println("Iniciando o servidor na porta "+ conf.Port)
	http.ListenAndServe("127.0.0.1:"+conf.Port, router)
}