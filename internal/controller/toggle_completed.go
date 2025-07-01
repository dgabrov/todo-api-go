package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type toggleCompleted struct {
	db     *sql.DB
	config data.ServerConfig
}

func toggleCompletedController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &toggleCompleted{db, config}
}

func (h *toggleCompleted) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processToggleCompleted(r)

	writeResponse(w, res, err)
}

func (h *toggleCompleted) processToggleCompleted(r *http.Request) (data.CompletedData, error) {
	res := data.CompletedData{}
	ctx := r.Context()

	var input data.TodoItemIdData
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return res, err
	}

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return res, err
	}

	persons := session.Persons
	res, err = srv.ToggleCompleted(ctx, persons, input.TodoItemId)

	return res, err
}
