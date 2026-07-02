package controller

import (
	"database/sql"
	"net/http"

	"gitlab.com/dgb9/todo-api/internal/data"
)

type passwordPut struct {
	db     *sql.DB
	config data.ServerConfig
}

func passwordPutController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &passwordPut{db, config}
}

func (p passwordPut) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}
