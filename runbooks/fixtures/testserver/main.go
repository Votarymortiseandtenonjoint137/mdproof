// Tiny REST API for mdproof E2E testing.
// Endpoints: GET /health, GET/POST/DELETE /items.
package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	var (
		items  []Item
		mu     sync.Mutex
		nextID = 1
	)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			mu.Lock()
			out := make([]Item, len(items))
			copy(out, items)
			mu.Unlock()
			json.NewEncoder(w).Encode(out)
		case "POST":
			var input struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
				return
			}
			mu.Lock()
			item := Item{ID: nextID, Name: input.Name}
			nextID++
			items = append(items, item)
			mu.Unlock()
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(item)
		case "DELETE":
			mu.Lock()
			items = nil
			nextID = 1
			mu.Unlock()
			w.WriteHeader(204)
		}
	})

	http.ListenAndServe(":18080", nil)
}
