package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type multipleTodo struct {
	db     *sql.DB
	config data.ServerConfig
}

func multipleTodoController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &multipleTodo{db, config}
}

func (h *multipleTodo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processMultipleTodo(r)

	writeResponse(w, res, err)
}

func (h *multipleTodo) processMultipleTodo(r *http.Request) ([]data.TodoItem, error) {
	ctx := r.Context()

	var input data.MultipleTodoData
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

	return srv.MultipleTodo(ctx, persons, input)
}
