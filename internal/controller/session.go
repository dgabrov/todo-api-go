package controller

import (
	"context"
	"errors"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
	"net/http"
	"strings"
)

type SessionRetriever interface {
	GetSessionByToken(ctx context.Context, token string) (data.Session, error)
}

func getSession(ctx context.Context, r *http.Request, retriever SessionRetriever) (data.Session, error) {
	res := data.Session{}
	token := ""

	// get the token from either the authorization header or from the cookie
	header := r.Header.Get("Authorization")
	header = strings.TrimSpace(header)

	if len(header) > 7 && strings.HasPrefix(strings.ToLower(header), "bearer ") {
		token = header[7:]
	}

	if len(token) == 0 {
		// ok try in the cookies
		cookie, err := r.Cookie(cst.TokenCookieName)

		if err == nil {
			token = cookie.Value
		}
	}

	// now get the session if possible
	var err error
	if len(token) > 0 {
		res, err = retriever.GetSessionByToken(ctx, token)

		if err != nil {
			return res, err
		}

		// this is the success path
		return res, nil
	}

	// with the token, get
	return res, errors.New("token not found or session not found for provided token")
}
