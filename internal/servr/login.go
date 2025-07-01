package servr

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr/dao"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

func (h *server) AddLogin(ctx context.Context, session data.Session, loginData data.LoginData) (data.Person, error) {
	res := data.Person{}

	err := checkLoginData(loginData)
	if err != nil {
		return res, err
	}

	// call the dependency
	auth, err := callIdp(h.config.AuthServerUrl, loginData)
	if err != nil {
		return res, err
	}

	// check to see if it has the expected right
	if !slices.Contains(auth.Rights, cst.AccessRightNeeded) {
		return res, errors.New("you don't have the right to connect to this application")
	}

	tx, err := begin(ctx, h.db)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	providedId := auth.Id
	name := auth.Name
	login := auth.Login

	person, err := addOrUpdatePerson(ctx, tx, providedId, name, login)
	if err != nil {
		return res, err
	}

	// get the person id and see if it is already
	if slices.Contains(session.Persons, person.PersonId) {
		return res, fmt.Errorf("user %s already logged in", login)
	}

	// all good, add person_id to the session_person for the given session
	err = dao.AddPersonForSession(ctx, tx, session.SessionId, person.PersonId)

	err = tx.Commit()

	res.PersonId = person.PersonId
	res.Name = person.Name
	res.Login = person.Login
	res.ProvidedId = person.ProvidedId

	return res, err
}

func (h *server) Login(ctx context.Context, loginData data.LoginData) (data.TokenPersonData, error) {
	res := data.TokenPersonData{}

	err := checkLoginData(loginData)
	if err != nil {
		return res, err
	}

	// call the dependency
	auth, err := callIdp(h.config.AuthServerUrl, loginData)
	if err != nil {
		return res, err
	}

	// check to see if it has the expected right
	if !slices.Contains(auth.Rights, cst.AccessRightNeeded) {
		return res, errors.New("you don't have the right to connect to this application")
	}

	tx, err := begin(ctx, h.db)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	providedId := auth.Id
	name := auth.Name
	login := auth.Login

	person, err := addOrUpdatePerson(ctx, tx, providedId, name, login)

	// we got the person, now creating the session
	token := uuid.New().String()
	newSessionId := uuid.New().String()
	session := data.Session{
		SessionId: newSessionId,
		Updated:   time.Now(),
		Token:     token,
		Expired:   false,
		Persons:   []string{person.PersonId},
	}

	err = dao.AddSession(ctx, tx, session)
	if err != nil {
		return res, err
	}

	res.PersonId = person.PersonId
	res.ProvidedId = person.ProvidedId
	res.Login = person.Login
	res.Name = person.Name
	res.Token = token

	err = dao.AddSessionPersons(ctx, tx, newSessionId, session.Persons)

	err = tx.Commit()

	return res, err
}

func callIdp(url string, login data.LoginData) (data.LoginAuthData, error) {
	var res data.LoginAuthData

	jsonBytes, err := json.Marshal(login)
	if err != nil {
		return res, err
	}
	buffer := bytes.NewReader(jsonBytes)

	resp, err := http.Post(url, cst.ContentTypeJson, buffer)
	if err != nil {
		return res, err
	}

	if resp.StatusCode >= 400 {
		bytes, er := io.ReadAll(resp.Body)
		if er != nil {
			return res, er
		}

		return res, errors.New(string(bytes))
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	return res, err
}

func checkLoginData(loginData data.LoginData) error {
	login := strings.TrimSpace(loginData.Login)
	if len(login) == 0 {
		return errors.New("login is mandatory and it seems to be missing")
	}

	password := strings.TrimSpace(loginData.Password)

	if len(password) == 0 {
		return errors.New("empty passwords are no longer valid, please provide the password")
	}

	return nil
}

func addOrUpdatePerson(ctx context.Context, tx *sql.Tx, providedId string, name string, login string) (data.Person, error) {
	person, err := dao.GetPersonByProvidedId(ctx, tx, providedId)
	if err != nil {
		return person, err
	}

	if len(person.PersonId) == 0 {
		// crete th
		person.PersonId = uuid.New().String()
		person.Name = name
		person.Login = login
		person.ProvidedId = providedId

		err = dao.AddPerson(ctx, tx, person)
		if err != nil {
			return person, err
		}
	} else {
		person.Name = name
		person.Login = login

		err = dao.UpdatePerson(ctx, tx, person)
		if err != nil {
			return person, err
		}
	}

	return person, nil
}

func (h *server) RemoveLogin(ctx context.Context, session data.Session, personId string) error {
	// person id must be filled out
	if len(strings.TrimSpace(personId)) == 0 {
		return errors.New("person id not filled out")
	}

	// person id must be part of the login values
	if !slices.Contains(session.Persons, personId) {
		return fmt.Errorf("person with id %s not logged in", personId)
	}

	// you cannot remove the last person from the session
	if len(session.Persons) == 1 {
		return errors.New("you cannot remove the only logged in person in the current session")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(tx)

	// proceed
	err = dao.RemoveSessionPerson(ctx, tx, session.SessionId, personId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *server) Logout(ctx context.Context, sessionId string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(tx)

	err = dao.LogoutSession(ctx, tx, sessionId)
	if err != nil {
		return err
	}

	return tx.Commit()
}
