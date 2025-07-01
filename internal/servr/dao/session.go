package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/dgb9/todo-api/internal/data"
	"time"
)

func AddSession(ctx context.Context, tx *sql.Tx, session data.Session) error {
	sql := "insert into session (session_id, updated_dt, token, expired_ind) values (?, ?, ?, ?)"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, session.SessionId, session.Updated, session.Token, convertBool(session.Expired))
	return err
}

func AddSessionPersons(ctx context.Context, tx *sql.Tx, sessionId string, persons []string) error {
	sql := "insert into session_person (session_person_id, session_id, person_id) values (?, ?, ?)"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	for _, person := range persons {
		newId := uuid.New().String()
		_, err = stat.ExecContext(ctx, newId, sessionId, person)

		if err != nil {
			return err
		}
	}

	return nil
}

func GetSessionByToken(ctx context.Context, tx *sql.Tx, token string) (data.Session, error) {
	res := data.Session{}
	sql := `select s.session_id, s.updated_dt, s.token, s.expired_ind, p.person_id 
				from session s left outer join session_person p on s.session_id = p.session_id 
				where s.token = ?`
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return res, err
	}

	rs, err := stat.QueryContext(ctx, token)
	if err != nil {
		return res, err
	}

	defer rs.Close()

	found := false
	var sessionId string
	var updatedDt time.Time
	var tk string
	var expiredInd string
	var personId string

	for rs.Next() {
		err = rs.Scan(&sessionId, &updatedDt, &tk, &expiredInd, &personId)

		if err != nil {
			return res, err
		}

		if !found {
			found = true

			res.SessionId = sessionId
			res.Updated = updatedDt
			res.Token = tk
			res.Expired = convertString(expiredInd)
		}

		res.Persons = append(res.Persons, personId)
	}

	if !found {
		return res, errors.New("token not found")
	}

	return res, nil
}

func AddPersonForSession(ctx context.Context, tx *sql.Tx, sessionId string, personId string) error {
	sql := "insert into session_person (session_person_id, session_id, person_id) values (?, ?, ?)"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, uuid.New().String(), sessionId, personId)

	return err
}

func RemoveSessionPerson(ctx context.Context, tx *sql.Tx, sessionId string, personId string) error {
	sql := "delete from session_person where session_id = ? and person_id = ?"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, sessionId, personId)

	return err
}

func LogoutSession(ctx context.Context, tx *sql.Tx, sessionId string) error {
	sql := "update session set expired_ind = 'Y' where session_id = ?"
	stat, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, sessionId)

	return err
}
