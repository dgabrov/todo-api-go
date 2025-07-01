package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type searchItem struct {
	db     *sql.DB
	config data.ServerConfig
}

func searchItemController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &searchItem{db, config}
}

func (h *searchItem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processSearchItem(r)

	writeResponse(w, res, err)
}

func (h *searchItem) processSearchItem(r *http.Request) ([]data.CompleteItemData, error) {
	ctx := r.Context()

	var input data.SearchData
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

	return srv.SearchItems(ctx, persons, input.Search)
}
