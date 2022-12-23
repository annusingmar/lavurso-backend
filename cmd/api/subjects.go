package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) listAllSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := app.models.Subjects.AllSubjects()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"subjects": subjects})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createSubject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	subject := &model.Subjects{
		Name: &input.Name,
	}

	v.Check(*subject.Name != "", "name", "must be provided")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.models.Subjects.InsertSubject(subject)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}

func (app *application) updateSubject(w http.ResponseWriter, r *http.Request) {
	subjectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if subjectID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchSubject.Error())
		return
	}

	subject, err := app.models.Subjects.GetSubjectByID(subjectID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSubject):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		Name *string `json:"name"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name != nil {
		subject.Name = input.Name
	}

	v := validator.NewValidator()

	v.Check(*subject.Name != "", "name", "must be provided")
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.models.Subjects.UpdateSubject(subject)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
