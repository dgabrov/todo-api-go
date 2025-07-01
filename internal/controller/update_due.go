package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type updateDue struct {
	db     *sql.DB
	config data.ServerConfig
}

func updateDueController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &updateDue{db, config}
}

func (h *updateDue) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processUpdateDue(r)
	writeResponse(w, res, err)
}

func (h *updateDue) processUpdateDue(r *http.Request) (bool, error) {
	ctx := r.Context()

	var input data.UpdateDueData
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return false, err
	}

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return false, err
	}

	persons := session.Persons
	err = srv.UpdateDue(ctx, persons, input)
	if err != nil {
		return false, err
	}

	return true, nil
}
