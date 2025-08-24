package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis/v2"
)

var (
	masterPool *simpleredis.ConnectionPool
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
}

func ListRangeHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	list := simpleredis.NewList(masterPool, key)

	members, err := list.GetAll()
	if err != nil {
		writeError(w, fmt.Errorf("failed to get list: %v", err), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(members); err != nil {
		writeError(w, fmt.Errorf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func ListPushHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	value := mux.Vars(r)["value"]
	list := simpleredis.NewList(masterPool, key)

	if err := list.Add(value); err != nil {
		writeError(w, fmt.Errorf("failed to add to list: %v", err), http.StatusServiceUnavailable)
		return
	}

	ListRangeHandler(w, r)
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	info, err := masterPool.Get(0).Do("INFO")
	if err != nil {
		writeError(w, fmt.Errorf("failed to get Redis info: %v", err), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(info.([]byte))
}

func EnvHandler(w http.ResponseWriter, r *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.SplitN(item, "=", 2)
		if len(splits) == 2 {
			environment[splits[0]] = splits[1]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(environment); err != nil {
		writeError(w, fmt.Errorf("failed to encode environment: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("local machine envs")
}

func initializeRedisConnection(host string, maxRetries int) (*simpleredis.ConnectionPool, error) {
	var pool *simpleredis.ConnectionPool
	var err error

	for i := 0; i < maxRetries; i++ {
		pool = simpleredis.NewConnectionPoolHost(host)
		_, err = pool.Get(0).Do("PING")
		if err == nil {
			return pool, nil
		}

		log.Printf("Failed to connect to Redis at %s (attempt %d/%d): %v", host, i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis at %s after %d attempts: %v", host, maxRetries, err)
}

func main() {
	demoVal := os.Getenv("TEST_ENV")

	if demoVal != "" {
		log.Printf("TEST_ENV exists and is %s!\n", demoVal)
	} else {
		log.Println("TEST_ENV does not exist.")
	}

	const maxRetries = 5

	var err error
	masterPool, err = initializeRedisConnection("redis-master:6379", maxRetries)
	if err != nil {
		log.Fatalf("Failed to initialize Redis master connection: %v", err)
	}
	defer masterPool.Close()

	r := mux.NewRouter()
	r.HandleFunc("/lrange/{key}", ListRangeHandler).Methods("GET")
	r.HandleFunc("/rpush/{key}/{value}", ListPushHandler).Methods("GET")
	r.HandleFunc("/info", InfoHandler).Methods("GET")
	r.HandleFunc("/env", EnvHandler).Methods("GET")

	log.Println("Starting backend API server on :3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
