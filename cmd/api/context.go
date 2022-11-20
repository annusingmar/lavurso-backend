package main

import (
	"context"
	"net/http"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

type lavursoContextKey string

func (app *application) setUserForContext(user *data.NUser, r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), lavursoContextKey("user"), user)
	return r.WithContext(ctx)
}

func (app *application) getUserFromContext(r *http.Request) *data.NUser {
	user, ok := r.Context().Value(lavursoContextKey("user")).(*data.NUser)
	if !ok {
		return nil
	}

	return user
}
