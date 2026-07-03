package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
)

type passwordPut struct {
	db     *sql.DB
	config data.ServerConfig
}

func passwordPutController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &passwordPut{db, config}
}

func (p passwordPut) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	err := p.processPasswordPut(request)

	writeResponse(writer, struct{}{}, err)
}

func (p passwordPut) processPasswordPut(r *http.Request) error {
	ctx := r.Context()

	var bundle data.PasswordBundle
	if err := json.NewDecoder(r.Body).Decode(&bundle); err != nil {
		return err
	}

	srv := servr.GetServr(p.db, p.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return err
	}

	if !slices.Contains(session.Persons, bundle.PersonId) {
		return errors.New("person processed is not among the logged in individuals")
	}

	return srv.UpdatePasswordByUserId(ctx, bundle.PersonId, bundle.Password, bundle.Payload)
}
