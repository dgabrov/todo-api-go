package servr

import (
	"errors"
	"gitlab.com/dgb9/todo-api/internal/data"
	"strings"
)

func checkItem(item data.ItemData) error {
	name := strings.TrimSpace(item.Name)
	if len(name) == 0 {
		return errors.New("please provide at least some item name")
	}

	return nil
}

func checkTodo(todo data.TodoItem) error {
	comments := strings.TrimSpace(todo.Comments)
	if len(comments) == 0 {
		return errors.New("please provide at least some todo comments")
	}

	if todo.Priority < 0 || todo.Priority > 100 {
		return errors.New("priority must be between 0 and 100 inclusive")
	}

	return nil
}
