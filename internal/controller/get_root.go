package controller

import (
	"database/sql"
	"gitlab.com/dgb9/todo-api/internal/data"
	"net/http"
)

type root struct {
	db     *sql.DB
	config data.ServerConfig
}

func rootController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &root{db, config}
}

func (h *root) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, "success", nil)
}
