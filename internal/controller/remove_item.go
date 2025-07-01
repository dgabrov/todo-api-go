package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type removeItem struct {
	db     *sql.DB
	config data.ServerConfig
}

func removeItemController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &removeItem{db, config}
}

func (h *removeItem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.processRemoveItem(r)
	res := err != nil

	writeResponse(w, res, err)
}

func (h *removeItem) processRemoveItem(r *http.Request) error {
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

	persons := session.Persons

	return srv.RemoveItems(ctx, persons, input.Ids)
}
