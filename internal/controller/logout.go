package controller

import (
	"database/sql"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type logout struct {
	db     *sql.DB
	config data.ServerConfig
}

func logoutController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &logout{db, config}
}

func (h *logout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.processLogout(r)

	writeResponse(w, true, err)
}

func (h *logout) processLogout(r *http.Request) error {
	ctx := r.Context()
	srv := servr.GetServr(h.db, h.config)

	session, err := getSession(ctx, r, srv)
	if err != nil {
		return err
	}

	// get the sessionId from the session and proceed
	sessionId := session.SessionId
	return srv.Logout(ctx, sessionId)
}
