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

type passwordPost struct {
	db     *sql.DB
	config data.ServerConfig
}

func passwordPostController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &passwordPost{db, config}
}

func (p *passwordPost) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	data, err := p.processPasswordPostController(request)

	writeResponse(writer, data, err)
}

func (p *passwordPost) processPasswordPostController(r *http.Request) (data.PasswordData, error) {
	res := data.PasswordData{}

	ctx := r.Context()
	var input data.PasswordPostInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return res, err
	}

	srv := servr.GetServr(p.db, p.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return res, err
	}

	persons := session.Persons
	// check the person id that is passed is one of the logged in persons, or else trigger an error

	if !slices.Contains(persons, input.PersonId) {
		return res, errors.New("person processed is not among the logged in individuals")
	}

	return srv.GetPasswordByUserId(ctx, input.PersonId, input.Password)
}
