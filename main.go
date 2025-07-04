package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TodoItem struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

var db *pgx.Conn

func main() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}
	defer db.Close(context.Background())

	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/", todoHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllTodos(w, r)
	case http.MethodPost:
		createTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/todos/"):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTodoByID(w, id)
	case http.MethodPut:
		updateTodo(w, r, id)
	case http.MethodDelete:
		softDeleteTodo(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAllTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT id, title, description, completed, deleted, created_at, updated_at FROM todos WHERE deleted = false")
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []TodoItem
	for rows.Next() {
		var t TodoItem
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Deleted, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	respondJSON(w, http.StatusOK, todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t TodoItem
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	t.ID = uuid.New()
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt

	_, err = db.Exec(context.Background(), `
        INSERT INTO todos (id, title, description, completed, deleted, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, t.ID, t.Title, t.Description, t.Completed, false, t.CreatedAt, t.UpdatedAt)

	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, t)
}

func getTodoByID(w http.ResponseWriter, id uuid.UUID) {
	var t TodoItem
	err := db.QueryRow(context.Background(), `
        SELECT id, title, description, completed, deleted, created_at, updated_at
        FROM todos WHERE id = $1 AND deleted = false
    `, id).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Deleted, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, t)
}

func updateTodo(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var t TodoItem
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	t.ID = id
	t.UpdatedAt = time.Now().UTC()

	result, err := db.Exec(context.Background(), `
        UPDATE todos SET title=$1, description=$2, completed=$3, updated_at=$4 WHERE id=$5 AND deleted = false
    `, t.Title, t.Description, t.Completed, t.UpdatedAt, id)

	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, "Update failed or not found", http.StatusNotFound)
		return
	}

	err = db.QueryRow(context.Background(), `
        SELECT created_at
        FROM todos WHERE id = $1 AND deleted = false
    `, id).Scan(&t.CreatedAt)

	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, t)
}

func softDeleteTodo(w http.ResponseWriter, id uuid.UUID) {
	result, err := db.Exec(context.Background(), `
        UPDATE todos SET deleted=true, updated_at=$1 WHERE id=$2
    `, time.Now().UTC(), id)

	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, "Delete failed or not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusNoContent, nil) // 204 No Content
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
