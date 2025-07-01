package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type searchTodo struct {
	db     *sql.DB
	config data.ServerConfig
}

func searchTodoController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &searchTodo{db, config}
}

func (h *searchTodo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processSearchTodo(r)

	writeResponse(w, res, err)
}

func (h *searchTodo) processSearchTodo(r *http.Request) ([]data.TodoItem, error) {
	ctx := r.Context()
	var input data.TodoSearch
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		return nil, err
	}

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return nil, err
	}

	persons := session.Persons

	// call the service there
	res, err := srv.SearchTodo(ctx, persons, input)
	return res, err
}
