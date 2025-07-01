package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type bulkUpdate struct {
	db     *sql.DB
	config data.ServerConfig
}

func bulkUpdateController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &bulkUpdate{db, config}
}

func (h *bulkUpdate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processBulkUpdate(r)

	writeResponse(w, res, err)
}

func (h *bulkUpdate) processBulkUpdate(r *http.Request) ([]data.TodoItem, error) {
	ctx := r.Context()
	var input data.BulkUpdateData

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

	return srv.BulkUpdateTodo(ctx, persons, input)
}
