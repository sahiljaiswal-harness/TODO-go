package main

import (
	"context"
	"errors"
	"os"
	"log"
	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
)

var db *pgx.Conn

func InitPostgres() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}
}

// Get all todos (optionally filter by completed)
func DBGetAllTodos(ctx context.Context, completedParam string) ([]TodoItem, error) {
	query := "SELECT id, title, description, completed, deleted, created_at, updated_at FROM todos WHERE deleted = false"
	if completedParam != "" {
		query += " AND completed = " + completedParam
	}
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []TodoItem
	for rows.Next() {
		var t TodoItem
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Deleted, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

// Create a new todo
func DBCreateTodo(ctx context.Context, t *TodoItem) error {
	_, err := db.Exec(ctx, `
        INSERT INTO todos (id, title, description, completed, deleted, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, t.ID, t.Title, t.Description, t.Completed, false, t.CreatedAt, t.UpdatedAt)
	return err
}

// Get todo by ID
func DBGetTodoByID(ctx context.Context, id uuid.UUID) (*TodoItem, error) {
	var t TodoItem
	err := db.QueryRow(ctx, `
        SELECT id, title, description, completed, deleted, created_at, updated_at
        FROM todos WHERE id = $1 AND deleted = false
    `, id).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Deleted, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update a todo fully
func DBUpdateTodo(ctx context.Context, t *TodoItem) error {
	result, err := db.Exec(ctx, `
        UPDATE todos SET title=$1, description=$2, completed=$3, updated_at=$4 WHERE id=$5 AND deleted = false
    `, t.Title, t.Description, t.Completed, t.UpdatedAt, t.ID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

// Patch a todo (partial update)
func DBPatchTodo(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return errors.New("no fields to update")
	}
	query := "UPDATE todos SET "
	args := []interface{}{}
	idx := 1
	for k, v := range fields {
		if idx > 1 {
			query += ", "
		}
		query += k + "=$" + string(rune(idx+'0'-1))
		args = append(args, v)
		idx++
	}
	query += ", updated_at=$" + string(rune(idx+'0'-1)) + " WHERE id=$" + string(rune(idx+'0')) + " AND deleted = false"
	args = append(args, uuid.UUID(id),)
	args = append(args, uuid.UUID(id))
	_, err := db.Exec(ctx, query, args...)
	return err
}

// Soft delete a todo
func DBSoftDeleteTodo(ctx context.Context, id uuid.UUID) error {
	result, err := db.Exec(ctx, `UPDATE todos SET deleted = true WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}