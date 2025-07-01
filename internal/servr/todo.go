package servr

import (
	"context"
	"errors"
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr/dao"
	"slices"
)

func (h *server) UpdatePriority(ctx context.Context, session data.Session, priority int, todoItemId string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	// see if it exists and if belongs to current user
	persons, err := dao.GetPersonsByTodoIds(ctx, tx, []string{todoItemId})
	if err != nil {
		return err
	}

	if len(persons) == 0 {
		return fmt.Errorf("cannot find todo item id %s in db", todoItemId)
	}

	loggedInPersons := session.Persons
	for _, personId := range persons {
		if !slices.Contains(loggedInPersons, personId) {
			return errors.New("the specified todo item does not belong to any of the logged in users")
		}
	}

	if priority < cst.PriorityMinimum || priority > cst.PriorityMaximum {
		return fmt.Errorf("priority must be between %d and %d inclusive", cst.PriorityMinimum, cst.PriorityMaximum)
	}

	err = dao.UpdatePriority(ctx, tx, todoItemId, priority)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *server) ToggleCompleted(ctx context.Context, persons []string, todoItemId string) (data.CompletedData, error) {
	res := data.CompletedData{}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	// check todoItemId exists and it belongs to one of the logged in persons
	owners, err := dao.GetPersonsByTodoIds(ctx, tx, []string{todoItemId})
	if err != nil {
		return res, err
	}

	if len(owners) == 0 {
		return res, fmt.Errorf("cannot find todo item with the id %s", todoItemId)
	}

	for _, personId := range owners {
		if !slices.Contains(persons, personId) {
			return res, errors.New("the todo item does not belong to any of the logged in users")
		}
	}

	// ok now we are good to proceed
	res, err = dao.ToggleCompleted(ctx, tx, todoItemId)

	err = tx.Commit()

	return res, err
}

func (h *server) RemoveTodo(ctx context.Context, persons []string, ids []string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(tx)

	personIds, err := dao.GetPersonsByTodoIds(ctx, tx, ids)
	if err != nil {
		return err
	}

	for _, pid := range personIds {
		if !slices.Contains(persons, pid) {
			return errors.New("at least one of the todo do not belong to any of the logged in user")
		}
	}

	// now remove the entry
	err = dao.RemoveTodo(ctx, tx, ids)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *server) GetTodo(ctx context.Context, persons []string, todoItemId string) (data.TodoItem, error) {
	res := data.TodoItem{}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer rollback(tx)

	res, err = dao.GetTodoItem(ctx, tx, todoItemId)
	if err != nil {
		return res, err
	}

	if len(res.TodoItemId) > 0 {
		// check the person
		owner := res.PersonId
		if !slices.Contains(persons, owner) {
			return data.TodoItem{}, errors.New("the todo item does not belong to any of the logged in users")
		}
	}

	err = tx.Commit()
	if err != nil {
		return data.TodoItem{}, err
	}

	return res, nil
}

func (h *server) UpdateDue(ctx context.Context, persons []string, input data.UpdateDueData) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	// check the entry exists
	item, err := dao.GetTodoItem(ctx, tx, input.TodoItemId)
	if err != nil {
		return err
	}

	if len(item.TodoItemId) == 0 {
		return errors.New("todo item does not exist")
	}

	// check the entry belongs to whoever it should belong
	personId := item.PersonId
	if !slices.Contains(persons, personId) {
		return errors.New("the item does not belong to any of the logged in persons")
	}

	// fill out the info
	err = dao.UpdateDue(ctx, tx, input)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *server) UpdateTodo(ctx context.Context, persons []string, input data.UpdateTodoItem) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(tx)

	// get the person id and see if it belongs to any of the logged in users
	todo := input.Todo
	personId := todo.PersonId
	if !slices.Contains(persons, personId) {
		return errors.New("the provided todo item does not belong to any of the logged in users")
	}

	err = checkTodo(todo)
	if err != nil {
		return err
	}
	var item data.TodoItem

	if input.Adding {
		err = dao.AddTodoItem(ctx, tx, todo)
		if err != nil {
			return err
		}
	} else {
		// check first that it  does exist
		item, err = dao.GetTodoItem(ctx, tx, todo.TodoItemId)
		if err != nil {
			return err
		}

		itemId := item.TodoItemId
		if len(itemId) == 0 {
			return fmt.Errorf("cannot find the todo item with id: %s", itemId)
		}

		err = dao.UpdateTodoItem(ctx, tx, todo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (h *server) SearchTodo(ctx context.Context, persons []string, input data.TodoSearch) ([]data.TodoItem, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer rollback(tx)

	res, err := dao.SearchTodo(ctx, tx, persons, input)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()

	return res, err
}

func (h *server) MultipleTodo(ctx context.Context, persons []string, input data.MultipleTodoData) ([]data.TodoItem, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer rollback(tx)

	// check the person belongs to persons
	if !slices.Contains(persons, input.PersonId) {
		return nil, errors.New("the person in the aggregate data is not one of the logged in persons")
	}

	// the priority must be between 0 and 100 inclusive
	if input.Priority < cst.PriorityMinimum || input.Priority > cst.PriorityMaximum {
		return nil, errors.New("priority must be between 0 and 100 inclusive")
	}

	res, err := dao.MultipleTodo(ctx, tx, input)
	if err != nil {
		return res, err
	}

	err = tx.Commit()

	return res, err
}

func (h *server) BulkUpdateTodo(ctx context.Context, persons []string, input data.BulkUpdateData) ([]data.TodoItem, error) {
	res := make([]data.TodoItem, 0)
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer rollback(tx)

	// the owner id must be one of the persons
	personIds, err := dao.GetPersonsByTodoIds(ctx, tx, input.TodoIds)

	for _, personId := range personIds {
		if !slices.Contains(persons, personId) {
			return nil, errors.New("at least one todo item does not belong to any of the logged in user")
		}
	}

	if !slices.Contains(persons, input.OwnerId) && input.SelectedOwner {
		return nil, errors.New("owner change is selected, however the selected owner is not one of the logged in users")
	}

	err = dao.BulkUpdate(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	res, err = dao.GetTodoItems(ctx, tx, input.TodoIds)

	err = tx.Commit()

	return res, err
}
