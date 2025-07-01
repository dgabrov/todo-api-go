package controller

import (
	"database/sql"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type flaggedItems struct {
	db     *sql.DB
	config data.ServerConfig
}

func flaggedItemsController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &flaggedItems{db, config}
}

func (h *flaggedItems) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processFlaggedItems(r)

	writeResponse(w, res, err)
}

func (h *flaggedItems) processFlaggedItems(r *http.Request) ([]data.CompleteItemData, error) {
	ctx := r.Context()

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return nil, err
	}
	persons := session.Persons

	res, err := srv.GetFlaggedItems(ctx, persons)

	return res, err
}
