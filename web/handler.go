package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/ragtag-archive/tasq/static"
)

const (
	MsgBadRequest    = "Bad request"
	MsgInternalError = "Internal error"
	MsgEmptyQueue    = "Queue is empty"
	MsgInsertError   = "Error inserting item"
	MsgCountError    = "Error getting count"
	MsgListError     = "Error getting tasks list"
	MsgPopError      = "Error popping item"
)

func Handler(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Method, r.URL.Path)

		queueName := r.URL.Path[1:]

		if queueName == "" {
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, static.IndexPage)
			return
		}

		ctx := r.Context()
		switch r.Method {
		case "PUT":
			// Read the PUT body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				httpError(w, http.StatusBadRequest, MsgBadRequest)
				return
			}

			// Insert
			if err := redisClient.ZIncrBy(ctx, queueName, 1, string(body)).Err(); err != nil {
				log.Println(err)
				httpError(w, http.StatusInternalServerError, MsgInsertError)
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
				httpError(w, http.StatusInternalServerError, MsgCountError)
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
				httpError(w, http.StatusInternalServerError, MsgListError)
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
				httpError(w, http.StatusInternalServerError, MsgPopError)
				return
			}

			// Check if there is an item
			if len(task) == 0 {
				httpError(w, http.StatusNotFound, MsgEmptyQueue)
				return
			}

			// Return the item
			data, ok := task[0].Member.(string)
			if !ok {
				httpError(w, http.StatusInternalServerError, MsgInternalError)
				return
			}

			httpOk(w, &GetResponse{Key: queueName + ":" + data, Data: data})
		}
	}
}
