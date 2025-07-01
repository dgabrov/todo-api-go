package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type removeLogin struct {
	db     *sql.DB
	config data.ServerConfig
}

func removeLoginController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &removeLogin{db, config}
}

func (h *removeLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.processRemoveLogin(r)

	writeResponse(w, res, err)
}

func (h *removeLogin) processRemoveLogin(r *http.Request) (int, error) {
	res := 0
	ctx := r.Context()

	var personIdData data.PersonIdData

	err := json.NewDecoder(r.Body).Decode(&personIdData)
	if err != nil {
		return res, err
	}

	serverInstance := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, serverInstance)

	if err != nil {
		return res, err
	}

	err = serverInstance.RemoveLogin(ctx, session, personIdData.PersonId)

	return res, err
}
