package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"
    "github.com/google/uuid"
)

type TodoItem struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    Deleted     bool      `json:"deleted"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

var (
    todos = make(map[string]*TodoItem)
    mu    sync.Mutex
)

func listTodos(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()
    var result []TodoItem
    for _, todo := range todos {
        if !todo.Deleted {
            result = append(result, *todo)
        }
    }
    json.NewEncoder(w).Encode(result)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
    var todo TodoItem
    json.NewDecoder(r.Body).Decode(&todo)
    todo.ID = uuid.New().String()
    todo.CreatedAt = time.Now()
    todo.UpdatedAt = time.Now()
    mu.Lock()
    todos[todo.ID] = &todo
    mu.Unlock()
    json.NewEncoder(w).Encode(todo)
}

func getTodo(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/todos/"):]
    mu.Lock()
    defer mu.Unlock()
    todo, exists := todos[id]
    if !exists || todo.Deleted {
        http.NotFound(w, r)
        return
    }
    json.NewEncoder(w).Encode(todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/todos/"):]
    var updated TodoItem
    json.NewDecoder(r.Body).Decode(&updated)

    mu.Lock()
    defer mu.Unlock()
    todo, exists := todos[id]
    if !exists || todo.Deleted {
        http.NotFound(w, r)
        return
    }
    todo.Title = updated.Title
    todo.Description = updated.Description
    todo.Completed = updated.Completed
    todo.UpdatedAt = time.Now()
    json.NewEncoder(w).Encode(todo)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/todos/"):]
    mu.Lock()
    defer mu.Unlock()
    todo, exists := todos[id]
    if !exists || todo.Deleted {
        http.NotFound(w, r)
        return
    }
    todo.Deleted = true
    todo.UpdatedAt = time.Now()
    w.WriteHeader(http.StatusNoContent)
}

func main() {
    http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            listTodos(w, r)
        case http.MethodPost:
            createTodo(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            getTodo(w, r)
        case http.MethodPut:
            updateTodo(w, r)
        case http.MethodDelete:
            deleteTodo(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    log.Println("Server started on :8080")
    http.ListenAndServe(":8080", nil)
}
