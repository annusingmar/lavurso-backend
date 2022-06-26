package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/julienschmidt/httprouter"
)

func (app *application) getClassForUser(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	userID, err := strconv.Atoi(params.ByName("id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
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

	if user.Role != data.Student {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	class, err := app.models.Classes.GetClassForUserID(user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoClassForUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"class": class})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getUsersInClass(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	classID, err := strconv.Atoi(params.ByName("id"))
	if classID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchClass.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(classID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	users, err := app.models.Classes.GetUsersForClassID(class.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"users": users})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}
