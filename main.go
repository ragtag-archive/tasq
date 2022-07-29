package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const index = `
                 _
                | |
 _ __ __ _  __ _| |_ __ _  __ _
| '__/ _' |/ _' | __/ _' |/ _' |
| | | (_| | (_| | || (_| | (_| |
|_|  \__,_|\__, |\__\__,_|\__, |
       _    __/ |          __/ |
      | |  |___/          |___/
      | |_ __ _ ___  __ _
      | __/ _' / __|/ _' |
      | || (_| \__ \ (_| |
       \__\__,_|___/\__, |
                       | |
                       |_|

A basic, easy to use task queue service.
----------------------------------------

PUT /:list
    Add the request body to the list.

    Example:
    curl -XPUT -d 'wowzers' https://tasq.url/test
    {
        "ok": true,
        "payload": {
            "key": "test:1606841828655"
        },
        "message": ""
    }


GET /:list
    List available task keys and total count in the
    specified list.

    Example:
    curl -XGET https://tasq.url/test
    {
        "ok": true,
        "payload": {
            "tasks": ["test:1606841828655"],
            "count": 1
        },
        "message": ""
    }


POST /:list
    Consume an item from the queue. Once consumed, the item
    will be removed from the list.

    Example:
    curl -XPOST https://tasq.url/test
    {
        "ok": true,
        "payload": {
            "key": "test:1606841828655",
            "data": "wowzers"
        },
        "message": ""
    }
`

type Response[T any] struct {
	Ok      bool   `json:"ok"`
	Payload T      `json:"payload"`
	Message string `json:"message"`
}

type Empty struct{}

type PutResponse struct {
	Key string `json:"key"`
}

type ListResponse struct {
	Tasks []string `json:"tasks"`
	Count int64    `json:"count"`
}

type GetResponse struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

var redisClient *redis.Client

func httpJson(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println(err)
	}
}

func httpOk[T any](w http.ResponseWriter, payload T) {
	httpJson(w, http.StatusOK, &Response[T]{
		Ok: true, Payload: payload, Message: "",
	})
}

func httpError(w http.ResponseWriter, status int, message string) {
	httpJson(w, http.StatusInternalServerError, &Response[Empty]{
		Ok: false, Payload: Empty{}, Message: message,
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path[1:]

	if queueName == "" {
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintf(w, index)
		return
	}

	ctx := r.Context()
	switch r.Method {
	case "PUT":
		// Read the PUT body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			httpError(w, http.StatusInternalServerError, "Error reading body")
			return
		}

		// Insert
		if err := redisClient.ZIncrBy(ctx, queueName, 1, string(body)).Err(); err != nil {
			log.Println(err)
			httpError(w, http.StatusInternalServerError, "Error adding to queue")
			return
		}

		// Return the key
		httpOk(w, &PutResponse{Key: queueName + ":" + string(body)})
		return

	case "GET":
		// Get the count
		count, err := redisClient.ZCard(ctx, queueName).Result()
		if err != nil {
			log.Println(err)
			httpError(w, http.StatusInternalServerError, "Error getting queue count")
			return
		}

		// Get the first 100 items, sorted by score descending
		tasks, err := redisClient.ZRangeArgs(ctx, redis.ZRangeArgs{
			Key:     queueName,
			ByScore: true,
			Rev:     true,
			Count:   100,
			Start:   "-inf",
			Stop:    "+inf",
		}).Result()

		if err != nil {
			log.Println(err)
			httpError(w, http.StatusInternalServerError, "Error getting queue items")
			return
		}

		keys := make([]string, len(tasks))
		for i, task := range tasks {
			keys[i] = queueName + ":" + task
		}

		// Return the list
		httpOk(w, &ListResponse{Tasks: keys, Count: count})
		return

	case "POST":
		// Pop the item with the highest score
		task, err := redisClient.ZPopMax(ctx, queueName, 1).Result()
		if err != nil {
			log.Println(err)
			httpError(w, http.StatusInternalServerError, "Error popping item")
			return
		}

		// Check if there is an item
		if len(task) == 0 {
			httpError(w, http.StatusNotFound, "Queue is empty")
			return
		}

		// Return the item
		data, ok := task[0].Member.(string)
		if !ok {
			httpError(w, http.StatusInternalServerError, "Error casting item")
			return
		}

		httpOk(w, &GetResponse{Key: queueName + ":" + data, Data: data})
	}
}

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
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       redisDatabase,
	})

	// Set up HTTP server
	log.Printf("Listening on %s", bindAddress)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(bindAddress, nil))
}
