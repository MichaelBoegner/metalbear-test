/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
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

func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	list := simpleredis.NewList(masterPool, key)
	
	members, err := list.GetAll()
	if err != nil {
		writeError(rw, fmt.Errorf("failed to get list: %v", err), http.StatusServiceUnavailable)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(members); err != nil {
		writeError(rw, fmt.Errorf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	list := simpleredis.NewList(masterPool, key)
	
	if err := list.Add(value); err != nil {
		writeError(rw, fmt.Errorf("failed to add to list: %v", err), http.StatusServiceUnavailable)
		return
	}
	
	ListRangeHandler(rw, req)
}

func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	info, err := masterPool.Get(0).Do("INFO")
	if err != nil {
		writeError(rw, fmt.Errorf("failed to get Redis info: %v", err), http.StatusServiceUnavailable)
		return
	}
	
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write(info.([]byte))
}

func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		environment[key] = val
	}

	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(environment); err != nil {
		writeError(rw, fmt.Errorf("failed to encode environment: %v", err), http.StatusInternalServerError)
		return
	}
}

func initializeRedisConnection(host string, maxRetries int) (*simpleredis.ConnectionPool, error) {
	var pool *simpleredis.ConnectionPool
	var err error
	
	for i := 0; i < maxRetries; i++ {
		pool = simpleredis.NewConnectionPoolHost(host)
		// Test the connection
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
	const maxRetries = 5
	
	var err error
	masterPool, err = initializeRedisConnection("redis-master:6379", maxRetries)
	if err != nil {
		log.Fatalf("Failed to initialize Redis master connection: %v", err)
	}
	defer masterPool.Close()
	
	r := mux.NewRouter()
	r.Path("/lrange/{key}").Methods("GET").HandlerFunc(ListRangeHandler)
	r.Path("/rpush/{key}/{value}").Methods("GET").HandlerFunc(ListPushHandler)
	r.Path("/info").Methods("GET").HandlerFunc(InfoHandler)
	r.Path("/env").Methods("GET").HandlerFunc(EnvHandler)

	n := negroni.Classic()
	n.UseHandler(r)
	
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", n); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
