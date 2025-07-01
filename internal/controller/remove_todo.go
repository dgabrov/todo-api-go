package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type removeTodo struct {
	db     *sql.DB
	config data.ServerConfig
}

func removeTodoController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &removeTodo{db, config}
}

func (h *removeTodo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.processRemoveTodo(r)

	writeResponse(w, true, err)
}

func (h *removeTodo) processRemoveTodo(r *http.Request) error {
	ctx := r.Context()
	var input data.IdsData
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		return err
	}

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)

	if err != nil {
		return err
	}

	return srv.RemoveTodo(ctx, session.Persons, input.Ids)
}
