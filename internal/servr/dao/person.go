package dao

import (
	"context"
	"database/sql"
	"gitlab.com/dgb9/todo-api/internal/data"
)

func GetPersonByProvidedId(ctx context.Context, tx *sql.Tx, providedId string) (data.Person, error) {
	res := data.Person{}

	sql := "select person_id, provided_id, person_name, login from person where provided_id = ?"
	statement, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return res, err
	}

	rs, err := statement.QueryContext(ctx, providedId)
	if err != nil {
		return res, err
	}

	defer rs.Close()

	if rs.Next() {
		err = rs.Scan(&res.PersonId, &res.ProvidedId, &res.Name, &res.Login)

		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func AddPerson(ctx context.Context, tx *sql.Tx, person data.Person) error {
	sql := "insert into person (person_id, provided_id, person_name, login) values (?,?,?,?)"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, person.PersonId, person.ProvidedId, person.Name, person.Login)

	return err
}

func UpdatePerson(ctx context.Context, tx *sql.Tx, person data.Person) error {
	sql := "update person set person_name = ?, login = ? where person_id = ?"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, person.Name, person.Login, person.PersonId)

	return err
}

func GetPersonPasswordData(ctx context.Context, tx *sql.Tx, personId string) (salt string, payload string, err error) {
	strSql := "select password_salt, password_payload from person where person_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return "", "", err
	}

	rs, err := stat.QueryContext(ctx, personId)
	if err != nil {
		return "", "", err
	}
	defer rs.Close()

	if rs.Next() {
		err = rs.Scan(&salt, &payload)
	}
	return salt, payload, err
}

func UpdatePersonPasswordData(ctx context.Context, tx *sql.Tx, personId string, salt string, payload string) error {
	strSql := "update person set password_salt = ?, password_payload = ? where person_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, salt, payload, personId)
	return err
}
