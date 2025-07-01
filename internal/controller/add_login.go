package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type addLogin struct {
	db     *sql.DB
	config data.ServerConfig
}

func addLoginController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &addLogin{db, config}
}

func (h *addLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	personData, err := h.processAddLogin(r)

	writeResponse(w, personData, err)
}

func (h *addLogin) processAddLogin(r *http.Request) (data.Person, error) {
	res := data.Person{}
	ctx := r.Context()

	var loginData data.LoginData
	err := json.NewDecoder(r.Body).Decode(&loginData)

	if err != nil {
		return res, err
	}

	servr := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, servr)
	if err != nil {
		return res, err
	}

	res, err = servr.AddLogin(ctx, session, loginData)
	return res, err
}
