package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/dgb9/todo-api/internal/data"
	"strings"
	"time"
)

func GetPersonsByTodoIds(ctx context.Context, tx *sql.Tx, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	strSql := "select distinct person_id from todo_item where todo_item_id in (_here_)"
	query := "?" + strings.Repeat(", ?", len(ids)-1)
	strSql = strings.Replace(strSql, "_here_", query, 1)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	var qr []any
	for _, crt := range ids {
		qr = append(qr, crt)
	}

	var id string
	var res []string
	rs, err := stat.QueryContext(ctx, qr...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	for rs.Next() {
		err = rs.Scan(&id)

		if err != nil {
			return nil, err
		}

		res = append(res, id)
	}

	return res, nil
}

func UpdatePriority(ctx context.Context, tx *sql.Tx, todoItemId string, priority int) error {
	strSql := "update todo_item set priority = ? where todo_item_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)

	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, priority, todoItemId)

	return err
}

func ToggleCompleted(ctx context.Context, tx *sql.Tx, todoItemId string) (data.CompletedData, error) {
	res := data.CompletedData{}
	strSql := "select completed_ind from todo_item where todo_item_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return res, err
	}

	rs, err := stat.QueryContext(ctx, todoItemId)
	if err != nil {
		return res, err
	}

	defer rs.Close()

	var completedInd string
	if rs.Next() {
		err = rs.Scan(&completedInd)

		if err != nil {
			return res, err
		}
	} else {
		return res, fmt.Errorf("still do not find the todo item with the id: %s", todoItemId)
	}

	rs.Close()

	newVal := !convertString(completedInd)

	strSql = "update todo_item set completed_ind = ? where todo_item_id = ?"
	stat1, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return res, err
	}

	_, err = stat1.ExecContext(ctx, convertBool(newVal), todoItemId)
	if err != nil {
		return res, err
	}

	res.Completed = newVal
	res.TodoItemId = todoItemId

	return res, nil
}

func RemoveTodo(ctx context.Context, tx *sql.Tx, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	strSql := "delete from todo_item where todo_item_id in (_here_)"
	query := "?" + strings.Repeat(", ?", len(ids)-1)
	strSql = strings.Replace(strSql, "_here_", query, 1)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	var qr []any
	for _, id := range ids {
		qr = append(qr, id)
	}

	_, err = stat.ExecContext(ctx, qr...)

	return err
}

func GetTodoItem(ctx context.Context, tx *sql.Tx, id string) (data.TodoItem, error) {
	res := data.TodoItem{}

	strSql := `select todo_item_id,
			   person_id,
			   comments,
			   project_cd,
			   context_cd,
			   priority,
			   added_dt,
			   due_dt,
			   completed_ind,
			   updated_dt
		from todo_item
		where todo_item_id = ?`

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return res, err
	}

	rs, err := stat.QueryContext(ctx, id)
	if err != nil {
		return res, err
	}

	defer rs.Close()

	if rs.Next() {
		res, err = loadTodoItem(rs)
		if err != nil {
			return res, err
		}

	}

	return res, nil
}

func loadTodoItem(rs *sql.Rows) (data.TodoItem, error) {
	var res data.TodoItem

	var todoItemId string
	var personId string
	var comments string
	var projectCd sql.NullString
	var contextCd sql.NullString
	var priority int
	var addedDt time.Time
	var dueDt sql.NullTime
	var completedInd sql.NullString
	var updatedDt sql.NullTime

	err := rs.Scan(&todoItemId, &personId, &comments, &projectCd, &contextCd,
		&priority, &addedDt, &dueDt, &completedInd, &updatedDt)

	if err != nil {
		return res, err
	}

	// fill out the result
	res.TodoItemId = todoItemId
	res.PersonId = personId
	res.Comments = comments

	if projectCd.Valid {
		res.ProjectCd = &projectCd.String
	}

	if contextCd.Valid {
		res.ContextCd = &contextCd.String
	}

	res.Priority = priority
	res.Added = addedDt

	if dueDt.Valid {
		res.Due = &dueDt.Time
	}

	res.Completed = false
	if completedInd.Valid {
		res.Completed = convertString(completedInd.String)
	}

	if updatedDt.Valid {
		res.Updated = updatedDt.Time
	}

	return res, nil
}

func UpdateDue(ctx context.Context, tx *sql.Tx, input data.UpdateDueData) error {
	id := input.TodoItemId
	due := input.Due

	strSql := "update todo_item set due_dt = ? where todo_item_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, due, id)
	return err
}

func AddTodoItem(ctx context.Context, tx *sql.Tx, todo data.TodoItem) error {
	strSql := `insert into todo_item (todo_item_id, 
                       					person_id, comments,
										project_cd, context_cd,
										priority, added_dt, due_dt,
										completed_ind, updated_dt)
						values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, todo.TodoItemId, todo.PersonId, todo.Comments,
		todo.ProjectCd, todo.ContextCd, todo.Priority, todo.Added, todo.Due,
		convertBool(todo.Completed), todo.Updated)
	return err
}

func UpdateTodoItem(ctx context.Context, tx *sql.Tx, todo data.TodoItem) error {
	strSql := `update todo_item
		set person_id     = ?,
			comments = ?,
			project_cd    = ?,
			context_cd    = ?,
			priority      = ?,
			added_dt      = ?,
			due_dt        = ?,
			completed_ind = ?,
			updated_dt    = ?
			where todo_item_id = ?`

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, todo.PersonId, todo.Comments, todo.ProjectCd,
		todo.ContextCd, todo.Priority, todo.Added, todo.Due,
		convertBool(todo.Completed), todo.Updated, todo.TodoItemId)

	return err
}

func SearchTodo(ctx context.Context, tx *sql.Tx, persons []string, input data.TodoSearch) ([]data.TodoItem, error) {
	if len(persons) == 0 {
		return nil, nil
	}

	// creating the sql script
	var params []any
	var sqlParts []string

	sqlParts = append(sqlParts, "select todo_item_id, person_id, comments, project_cd, context_cd, priority, added_dt, due_dt, completed_ind, updated_dt from todo_item")
	sqlParts = append(sqlParts, " where person_id in (? ", strings.Repeat(", ?", len(persons)-1), ")")
	for _, person := range persons {
		params = append(params, person)
	}

	if input.Completed != nil {
		sqlParts = append(sqlParts, " and completed_ind = ? ")
		params = append(params, convertBool(*input.Completed))
	}

	if len(input.Context) > 0 {
		var parts []string

		for _, ct := range input.Context {
			parts = append(parts, "context_cd like ? ")
			params = append(params, prepSearch(ct))
		}

		joined := strings.Join(parts, " or ")
		sqlParts = append(sqlParts, " and (", joined, " ) ")
	}

	if len(input.Project) > 0 {
		var parts []string

		for _, proj := range input.Project {
			parts = append(parts, "project_cd like ? ")
			params = append(params, prepSearch(proj))
		}

		joined := strings.Join(parts, " or ")
		sqlParts = append(sqlParts, " and (", joined, " ) ")
	}

	first := true
	if input.DueNull || len(input.DueInterval) > 0 {
		sqlParts = append(sqlParts, "and (")

		if input.DueNull {
			first = false
			sqlParts = append(sqlParts, " due_dt is null ")
		}

		for _, interval := range input.DueInterval {

			if interval.StartDate != nil || interval.EndDate != nil {
				if first {
					first = false
				} else {
					sqlParts = append(sqlParts, " or ")
				}

				sqlParts = append(sqlParts, "(")

				if interval.StartDate != nil {
					if interval.EndDate != nil {
						sqlParts = append(sqlParts, " due_dt >= ? and due_dt < ?")
						params = append(params, interval.StartDate, interval.EndDate)
					} else {
						sqlParts = append(sqlParts, " due_dt >= ? ")
						params = append(params, interval.StartDate)
					}
				} else {
					// end date automatically is not null in this situation
					sqlParts = append(sqlParts, " due_dt < ? ")
					params = append(params, interval.EndDate)
				}

				sqlParts = append(sqlParts, ")")
			}
		}

		sqlParts = append(sqlParts, ")")
	}

	// ok now the general
	if len(input.General) > 0 {
		for _, sr := range input.General {
			trimmed := strings.TrimSpace(sr)

			if len(trimmed) > 0 {
				sqlParts = append(sqlParts, " and comments like ?")
				params = append(params, prepSearch(trimmed))
			}
		}
	}

	sqlParts = append(sqlParts, "order by priority")

	strSql := strings.Join(sqlParts, " ")
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	rs, err := stat.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	res := make([]data.TodoItem, 0)

	for rs.Next() {
		var td data.TodoItem

		var completedInd string
		err = rs.Scan(&td.TodoItemId, &td.PersonId, &td.Comments, &td.ProjectCd, &td.ContextCd, &td.Priority, &td.Added, &td.Due, &completedInd, &td.Updated)
		if err != nil {
			return nil, err
		}
		td.Completed = convertString(completedInd)

		res = append(res, td)
	}

	return res, nil
}

func MultipleTodo(ctx context.Context, tx *sql.Tx, input data.MultipleTodoData) ([]data.TodoItem, error) {
	res := make([]data.TodoItem, 0)

	strSql := `insert into todo_item 
			(todo_item_id, person_id, comments, project_cd, context_cd, priority, added_dt, 
			 due_dt, completed_ind, updated_dt) 
			values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	comments := input.Comments
	var todoItem data.TodoItem
	for _, comment := range comments {
		trimmed := strings.TrimSpace(comment)

		if len(trimmed) > 0 {
			newId := uuid.New().String()

			currentDate := time.Now()
			_, err = stat.ExecContext(ctx, newId, input.PersonId, trimmed, input.ProjectCd, input.ContextCd, input.Priority, currentDate, input.Due, "N", currentDate)

			if err != nil {
				return nil, err
			}

			// get the newly added value
			todoItem, err = GetTodoItem(ctx, tx, newId)
			if err != nil {
				return nil, err
			}

			res = append(res, todoItem)
		}
	}

	return res, nil
}

func GetTodoItems(ctx context.Context, tx *sql.Tx, ids []string) ([]data.TodoItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var parts []string
	parts = append(parts, "select todo_item_id, person_id, comments, project_cd, context_cd, priority, added_dt, due_dt, completed_ind, updated_dt from todo_item where todo_item_id in (?", strings.Repeat(", ?", len(ids)-1), ")")
	strSql := strings.Join(parts, " ")

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	params := getStringParams(ids)

	rs, err := stat.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	res := make([]data.TodoItem, 0)
	for rs.Next() {
		item, err := loadTodoItem(rs)
		if err != nil {
			return nil, err
		}

		res = append(res, item)
	}

	return res, nil
}

func BulkUpdate(ctx context.Context, tx *sql.Tx, input data.BulkUpdateData) error {
	ids := input.TodoIds
	someSelected := input.SelectedOwner || input.SelectedContext || input.SelectedProject || input.SelectedDue || input.SelectedPriority

	if len(ids) == 0 || !someSelected {
		return nil
	}

	var parts []string
	var params []any

	parts = append(parts, "update todo_item set")
	first := true

	if input.SelectedOwner {
		first, parts = getFirst(first, parts, ",")

		parts = append(parts, "person_id = ?")
		params = append(params, input.OwnerId)
	}

	if input.SelectedContext {
		first, parts = getFirst(first, parts, ",")

		parts = append(parts, "context_cd = ?")
		params = append(params, input.Context)
	}

	if input.SelectedProject {
		first, parts = getFirst(first, parts, ",")

		parts = append(parts, "project_cd = ?")
		params = append(params, input.Project)
	}

	if input.SelectedDue {
		first, parts = getFirst(first, parts, ",")

		parts = append(parts, "due_dt = ?")
		params = append(params, input.Due)
	}

	if input.SelectedPriority {
		first, parts = getFirst(first, parts, ",")

		parts = append(parts, "priority = ?")
		params = append(params, input.Priority)
	}

	parts = append(parts, "where todo_item_id in (?", strings.Repeat(", ?", len(ids)-1), ")")
	idsStringParms := getStringParams(ids)
	params = append(params, idsStringParms...)

	strSql := strings.Join(parts, " ")
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, params...)
	return err
}
