package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type removeAttachment struct {
	db     *sql.DB
	config data.ServerConfig
}

func removeAttachmentController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &removeAttachment{db, config}
}

func (h *removeAttachment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processRemoveAttachment(r)

	writeResponse(w, res, err)
}

func (h *removeAttachment) processRemoveAttachment(r *http.Request) (bool, error) {
	ctx := r.Context()
	var input data.IdsData

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

	err = srv.RemoveAttachment(ctx, persons, input.Ids)

	return true, err
}
