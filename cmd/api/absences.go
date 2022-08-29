package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) excuseAbsenceForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	markID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if markID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchMark.Error())
		return
	}

	var input struct {
		Excuse string `json:"excuse"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	mark, err := app.models.Marks.GetMarkByID(markID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchMark):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if mark.Type != data.MarkAbsent {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchMark.Error())
		return
	}

	if *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOrParentOfStudent(mark.UserID, *sessionUser.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	at := time.Now().UTC()

	excuse := &data.Excuse{
		MarkID: &markID,
		Excuse: &input.Excuse,
		By:     &data.User{ID: sessionUser.ID},
		At:     &at,
	}

	v := validator.NewValidator()

	v.Check(*excuse.Excuse != "", "excuse", "must be provided")
	v.Check(*excuse.MarkID > 0, "absence_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	if mark.Excuse.MarkID != nil {
		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrAbsenceExcused.Error())
		return
	}

	err = app.models.Absences.InsertExcuse(excuse)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}

func (app *application) deleteExcuseForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	markID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if markID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchMark.Error())
		return
	}

	mark, err := app.models.Marks.GetMarkByID(markID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchMark):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if mark.Type != data.MarkAbsent {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchMark.Error())
		return
	}

	if *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOrParentOfStudent(mark.UserID, *sessionUser.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	err = app.models.Absences.DeleteExcuseByMarkID(mark.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}
