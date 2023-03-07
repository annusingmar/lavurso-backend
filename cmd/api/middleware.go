package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	ErrAuthenticationRequired = errors.New("authentication required")
)

func (app *application) authenticateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		splitHeader := strings.Split(authHeader, " ")
		if len(splitHeader) != 2 || splitHeader[0] != "Bearer" || len(splitHeader[1]) != 52 {
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

		err = app.models.Sessions.ExtendSession(*user.SessionID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		r = app.setUserForContext(user, r)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		if user == nil {
			app.writeErrorResponse(w, r, http.StatusUnauthorized, ErrAuthenticationRequired.Error())
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAdministrator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		if *user.Role != data.RoleAdministrator {
			app.notAllowed(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		if *user.Role != data.RoleTeacher && *user.Role != data.RoleAdministrator {
			app.notAllowed(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()

		log := &data.Log{
			Method: &r.Method,
			Target: helpers.ToPtr(r.URL.EscapedPath()),
			IP:     helpers.ToPtr(app.getIP(r)),
			At:     helpers.ToPtr(time.Now().UTC()),
		}

		user := app.getUserFromContext(r)
		if user != nil {
			log.UserID = &user.ID
			log.SessionID = user.SessionID
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		log.ResponseCode = helpers.ToPtr(ww.Status())
		log.Duration = helpers.ToPtr(int(time.Since(t1).Milliseconds()))

		app.models.Logs.InsertLog(log)
	})
}
