package web

import (
	"encoding/json"
	"log"
	"net/http"
)

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
	if message == "" {
		message = http.StatusText(status)
	}
	httpJson(w, status, &Response[Empty]{
		Ok: false, Payload: Empty{}, Message: message,
	})
}
