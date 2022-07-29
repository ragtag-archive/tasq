package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/ragtag-archive/tasq/web"
)

func main() {
	// Get configuration
	redisUrl := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDatabase := 0
	if dbStr := os.Getenv("REDIS_DATABASE"); dbStr != "" {
		redisDatabase, _ = strconv.Atoi(dbStr)
	}
	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = ":8080"
	}

	// Connect to redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       redisDatabase,
	})

	// Set up HTTP server
	log.Printf("Listening on %s", bindAddress)
	http.HandleFunc("/", web.Handler(redisClient))
	log.Fatal(http.ListenAndServe(bindAddress, nil))
}
