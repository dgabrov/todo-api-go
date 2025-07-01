package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type updateTodo struct {
	db     *sql.DB
	config data.ServerConfig
}

func updateTodoController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &updateTodo{db, config}
}

func (h *updateTodo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processUpdateTodo(r)

	writeResponse(w, res, err)
}

func (h *updateTodo) processUpdateTodo(r *http.Request) (int, error) {
	ctx := r.Context()
	srv := servr.GetServr(h.db, h.config)

	var input data.UpdateTodoItem
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		return 1, err
	}

	session, err := getSession(ctx, r, srv)
	if err != nil {
		return 1, err
	}

	persons := session.Persons

	// proceed
	err = srv.UpdateTodo(ctx, persons, input)
	if err != nil {
		return 1, err
	}

	return 0, nil
}
