package main

import (
	"context"
	"net/http"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

type lavursoContextKey string

func (app *application) setUserForContext(user *data.User, r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), lavursoContextKey("user"), user)
	return r.WithContext(ctx)
}

func (app *application) getUserFromContext(r *http.Request) *data.User {
	user, ok := r.Context().Value(lavursoContextKey("user")).(*data.User)
	if !ok {
		return nil
	}

	return user
}
