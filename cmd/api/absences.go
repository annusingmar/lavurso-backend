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

func (app *application) getAbsencesForStudent(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	absences, err := app.models.Absences.GetAbsenceMarksByUserID(user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"absences": absences})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) excuseAbsenceForStudent(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	var input struct {
		AbsenceMarkID int    `json:"absence_id"`
		Excuse        string `json:"excuse"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	by := 2
	at := time.Now().UTC()

	excuse := &data.AbsenceExcuse{
		AbsenceMarkID: &input.AbsenceMarkID,
		Excuse:        &input.Excuse,
		By:            &by, // to change
		At:            &at,
	}

	v := validator.NewValidator()

	v.Check(*excuse.Excuse != "", "excuse", "must be provided")
	v.Check(*excuse.AbsenceMarkID > 0, "absence_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	absence, err := app.models.Absences.GetAbsenceByMarkID(*excuse.AbsenceMarkID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchAbsence):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if absence.UserID != user.ID {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotValidAbsence.Error())
		return
	}

	if absence.AbsenceExcuses != nil {
		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrAbsenceExcused.Error())
		return
	}

	err = app.models.Absences.InsertExcuse(excuse)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"excuse": excuse})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

}

func (app *application) deleteAbsenceExcuseForStudent(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "sid"))
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

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	excuseID, err := strconv.Atoi(chi.URLParam(r, "eid"))
	if excuseID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchExcuse.Error())
		return
	}

	excuse, err := app.models.Absences.GetExcuseByID(excuseID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchExcuse):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	absence, err := app.models.Absences.GetAbsenceByMarkID(*excuse.AbsenceMarkID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	if absence.UserID != user.ID {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotValidAbsence.Error())
		return
	}

	err = app.models.Absences.DeleteExcuse(*excuse.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}
