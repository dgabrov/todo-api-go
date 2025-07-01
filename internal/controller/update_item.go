package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type updateItem struct {
	db     *sql.DB
	config data.ServerConfig
}

func updateItemController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &updateItem{db, config}
}

func (h *updateItem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processUpdateItem(r)

	writeResponse(w, res, err)
}

func (h *updateItem) processUpdateItem(r *http.Request) (data.ItemData, error) {
	res := data.ItemData{}
	ctx := r.Context()

	var input data.UpdateItemData
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

	return srv.UpdateItem(ctx, persons, input)
}
