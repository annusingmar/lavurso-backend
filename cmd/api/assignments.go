package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) getAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	assignmentID, err := strconv.Atoi(params.ByName("id"))
	if assignmentID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchAssignment.Error())
		return
	}

	assignment, err := app.models.Assignments.GetAssignmentByID(assignmentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchAssignment):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"assignment": assignment})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createAssignment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		JournalID   int       `json:"journal_id"`
		Description string    `json:"description"`
		Deadline    data.Date `json:"deadline"`
		Type        int       `json:"type"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	assignment := &data.Assignment{
		JournalID:   input.JournalID,
		Description: input.Description,
		Deadline:    input.Deadline,
		Type:        input.Type,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	if assignment.Deadline.Time.IsZero() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrInvalidDateFormat.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(assignment.JournalID > 0, "journal_id", "must be provided and valid")
	v.Check(assignment.Type == data.Homework || assignment.Type == data.Test, "type", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	journal, err := app.models.Journals.GetJournalByID(assignment.JournalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	err = app.models.Assignments.InsertAssignment(assignment)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"assignment": assignment})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) updateAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	assignmentID, err := strconv.Atoi(params.ByName("id"))
	if assignmentID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchAssignment.Error())
		return
	}

	assignment, err := app.models.Assignments.GetAssignmentByID(assignmentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchAssignment):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		Description *string    `json:"description"`
		Deadline    *data.Date `json:"deadline"`
		Type        *int       `json:"type"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Description != nil {
		assignment.Description = *input.Description
	}
	if input.Deadline != nil {
		assignment.Deadline = *input.Deadline
	}
	if input.Type != nil {
		assignment.Type = *input.Type
	}

	v := validator.NewValidator()

	v.Check(assignment.Type == data.Homework || assignment.Type == data.Test, "type", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	assignment.UpdatedAt = time.Now().UTC()

	err = app.models.Assignments.UpdateAssignment(assignment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"assignment": assignment})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	assignmentID, err := strconv.Atoi(params.ByName("id"))
	if assignmentID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchAssignment.Error())
		return
	}

	assignment, err := app.models.Assignments.GetAssignmentByID(assignmentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchAssignment):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Assignments.DeleteAssignment(assignment.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
