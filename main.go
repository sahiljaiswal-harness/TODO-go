package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func main() {
	InitPostgres()
	defer db.Close(context.Background())
	InitRedis()
	defer RDB.Close()

	r := mux.NewRouter()
	r.HandleFunc("/todos", todosHandler).Methods("GET", "POST")
	r.HandleFunc("/todos/{id}", todoHandler).Methods("GET", "PUT", "DELETE", "PATCH")

	// Set up channel to listen for interrupt or terminate signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start your server in a goroutine
	go func() {
		fmt.Println("Server running on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", r))
	}()

	// Wait for signal
	<-quit
	fmt.Println("Shutting down gracefully...")
}

func getAllTodos(w http.ResponseWriter, r *http.Request) {
	completedParam := r.URL.Query().Get("completed")
	cacheKey := "todos:list"
	if completedParam != "" {
		cacheKey += ":completed=" + completedParam
	}

	// Check cache first
	cached, err := RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		fmt.Println("serving from cache")
		w.Header().Set("X-Cache-Hit", "true")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cached))
		return
	}

	todos, err := DBGetAllTodos(ctx, completedParam)
	if err != nil {
		jsonError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(todos)
	RDB.Set(ctx, cacheKey, jsonData, 60*time.Second) // Set TTL 60s
	respondJSON(w, http.StatusOK, todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t TodoItem
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	t.ID = uuid.New()
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt
	if err := DBCreateTodo(ctx, &t); err != nil {
		jsonError(w, "DB insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Invalidate cache
	keys, _ := RDB.Keys(ctx, "todos:list*").Result()
	for _, key := range keys {
		RDB.Del(ctx, key)
	}
	respondJSON(w, http.StatusCreated, t)
}


func getTodoByID(w http.ResponseWriter, id uuid.UUID) {
	cacheKey := "todos:item:" + id.String()
	cached, err := RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		fmt.Println("serving todo from cache")
		w.Header().Set("X-Cache-Hit", "true")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}
	todo, err := DBGetTodoByID(ctx, id)
	if err != nil {
		jsonError(w, "Todo not found", http.StatusNotFound)
		return
	}
	jsonData, _ := json.Marshal(todo)
	RDB.Set(ctx, cacheKey, jsonData, 60*time.Second)
	respondJSON(w, http.StatusOK, todo)
}


func updateTodo(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var t TodoItem
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	t.ID = id
	t.UpdatedAt = time.Now().UTC()
	if err := DBUpdateTodo(ctx, &t); err != nil {
		jsonError(w, "Update failed or not found", http.StatusNotFound)
		return
	}
	todo, err := DBGetTodoByID(ctx, id)
	if err != nil {
		jsonError(w, "Todo not found", http.StatusNotFound)
		return
	}
	// Invalidate cache
	RDB.Del(ctx, "todos:item:"+id.String())
	keys, _ := RDB.Keys(ctx, "todos:list*").Result()
	for _, key := range keys {
		RDB.Del(ctx, key)
	}
	respondJSON(w, http.StatusOK, todo)
}

func patchTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid UUID", http.StatusBadRequest)
		return
	}
	var fields map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	fields["updated_at"] = time.Now().UTC()
	if err := DBPatchTodo(ctx, id, fields); err != nil {
		jsonError(w, "Patch failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Invalidate cache
	RDB.Del(ctx, "todos:item:"+id.String())
	keys, _ := RDB.Keys(ctx, "todos:list*").Result()
	for _, key := range keys {
		RDB.Del(ctx, key)
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "patched"})
}

func softDeleteTodo(w http.ResponseWriter, id uuid.UUID) {
	if err := DBSoftDeleteTodo(ctx, id); err != nil {
		jsonError(w, "Delete failed or not found", http.StatusNotFound)
		return
	}
	// Invalidate cache
	keys, _ := RDB.Keys(ctx, "todos:list*").Result()
	for _, key := range keys {
		RDB.Del(ctx, key)
	}
	RDB.Del(ctx, "todos:item:"+id.String())
	respondJSON(w, http.StatusNoContent, nil)
}
