package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/go-chi/chi/v5"
)

func (app *application) allSessionsForUser(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	switch *sessionUser.Role {
	case data.RoleAdministrator:
	default:
		if *sessionUser.ID != userID {
			app.notAllowed(w, r)
			return
		}
	}

	user, err := app.models.Users.GetUserByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	sessions, err := app.models.Sessions.GetSessionsByUserID(*user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"sessions": sessions})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeAllSessionsForUser(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	switch *sessionUser.Role {
	case data.RoleAdministrator:
	default:
		if *sessionUser.ID != userID {
			app.notAllowed(w, r)
			return
		}
	}

	user, err := app.models.Users.GetUserByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Sessions.RemoveAllSessionsByUserID(*user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeSession(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	sessionID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if sessionID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchSession.Error())
		return
	}

	session, err := app.models.Sessions.GetSessionByID(sessionID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSession):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	switch *sessionUser.Role {
	case data.RoleAdministrator:
	default:
		if *sessionUser.ID != *session.User.ID {
			app.notAllowed(w, r)
			return
		}
	}

	err = app.models.Sessions.RemoveSessionByID(session.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
