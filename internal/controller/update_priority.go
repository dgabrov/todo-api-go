package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type updatePriority struct {
	db     *sql.DB
	config data.ServerConfig
}

func updatePriorityController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &updatePriority{db, config}
}

func (h *updatePriority) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.processUpdatePriority(r)

	writeResponse(w, true, err)
}

func (h *updatePriority) processUpdatePriority(r *http.Request) error {
	ctx := r.Context()
	server := servr.GetServr(h.db, h.config)

	session, err := getSession(ctx, r, server)
	if err != nil {
		return err
	}

	var request data.UpdatePriorityData
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return err
	}

	return server.UpdatePriority(ctx, session, request.Priority, request.TodoItemId)
}
