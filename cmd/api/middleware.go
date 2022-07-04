package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

func (app *application) authenticateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		splitHeader := strings.Split(authHeader, " ")
		if len(splitHeader) != 2 || splitHeader[0] != "Bearer" || len(splitHeader[1]) != 26 {
			app.writeErrorResponse(w, r, http.StatusUnauthorized, data.ErrInvalidToken.Error())
			return
		}

		user, err := app.models.Users.GetUserBySessionToken(splitHeader[1])
		if err != nil {
			switch {
			case errors.Is(err, data.ErrInvalidToken):
				app.writeErrorResponse(w, r, http.StatusUnauthorized, err.Error())
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}

		err = app.models.Sessions.UpdateLastSeen(user.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		r = app.setUserForContext(user, r)
		next.ServeHTTP(w, r)
	})
}
