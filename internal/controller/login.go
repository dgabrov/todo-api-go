package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type login struct {
	db     *sql.DB
	config data.ServerConfig
}

func loginController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &login{db, config}
}

func (h *login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenData, err := h.processLogin(r)

	writeResponse(w, tokenData, err)
}

func (h *login) processLogin(r *http.Request) (data.TokenPersonData, error) {
	// get the login data from request and double check all good
	res := data.TokenPersonData{}
	ctx := r.Context()

	var loginData data.LoginData
	err := json.NewDecoder(r.Body).Decode(&loginData)

	if err != nil {
		return res, err
	}

	res, err = servr.GetServr(h.db, h.config).Login(ctx, loginData)

	return res, err
}
