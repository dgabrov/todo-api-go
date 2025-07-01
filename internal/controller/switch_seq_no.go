package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type switchSeqNo struct {
	db     *sql.DB
	config data.ServerConfig
}

func switchSeqNoController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &switchSeqNo{db, config}
}

func (h *switchSeqNo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	item, err := h.processSwitchSeqNo(r)

	writeResponse(w, item, err)
}

func (h *switchSeqNo) processSwitchSeqNo(r *http.Request) (data.CompleteItemData, error) {
	res := data.CompleteItemData{}
	var input []string
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return res, err
	}

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)

	if err != nil {
		return res, err
	}

	res, err = srv.SwitchSeqNo(ctx, session.Persons, input)

	return res, err
}
