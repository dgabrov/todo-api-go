package controller

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"net/http"
)

type retrieveTodo struct {
	db     *sql.DB
	config data.ServerConfig
}

func retrieveTodoController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &retrieveTodo{db, config}
}

func (h *retrieveTodo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	todo, err := h.processRetrieveTodo(r)
	writeResponse(w, todo, err)
}

func (h *retrieveTodo) processRetrieveTodo(r *http.Request) (data.TodoContainerData, error) {
	res := data.TodoContainerData{}
	ctx := r.Context()

	var input data.TodoItemIdData
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return res, err
	}

	// check session
	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return res, err
	}

	persons := session.Persons
	todoItemId := input.TodoItemId
	item, err := srv.GetTodo(ctx, persons, todoItemId)
	if err != nil {
		return res, err
	}

	if len(item.TodoItemId) > 0 {
		res.Todo = &item
	}

	return res, err
}
