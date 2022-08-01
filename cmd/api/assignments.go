package main

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) getAssignment(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		if journal.Teacher.ID != sessionUser.ID {
			app.notAllowed(w, r)
			return
		}
	case data.RoleParent:
		ok, err := app.models.Journals.DoesParentHaveChildInJournal(sessionUser.ID, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		ok, err := app.models.Journals.IsUserInJournal(sessionUser.ID, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"assignment": assignment})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createAssignment(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		JournalID   int       `json:"journal_id"`
		Description string    `json:"description"`
		Deadline    data.Date `json:"deadline"`
		Type        string    `json:"type"`
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
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Version:     1,
	}

	if assignment.Deadline.Time.IsZero() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrInvalidDateFormat.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(assignment.JournalID > 0, "journal_id", "must be provided and valid")
	v.Check(assignment.Type == data.AssignmentHomework || assignment.Type == data.AssignmentTest, "type", "must be provided and valid")

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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
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

	err = app.models.Journals.SetJournalLastUpdated(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) updateAssignment(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		Description *string    `json:"description"`
		Deadline    *data.Date `json:"deadline"`
		Type        *string    `json:"type"`
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

	v.Check(assignment.Type == data.AssignmentHomework || assignment.Type == data.AssignmentTest, "type", "must be provided and valid")

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

	err = app.models.Journals.SetJournalLastUpdated(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Assignments.DeleteAssignment(assignment.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Journals.SetJournalLastUpdated(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getAssignmentsForJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		if journal.Teacher.ID != sessionUser.ID {
			app.notAllowed(w, r)
			return
		}
	case data.RoleParent:
		ok, err := app.models.Journals.DoesParentHaveChildInJournal(sessionUser.ID, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		ok, err := app.models.Journals.IsUserInJournal(sessionUser.ID, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	assignments, err := app.models.Assignments.GetAssignmentsByJournalID(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"assignments": assignments})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getAssignmentsForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleParent:
		ok, err := app.models.Users.IsParentOfChild(sessionUser.ID, userID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		if sessionUser.ID != userID {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
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

	journals, err := app.models.Journals.GetJournalsByUserID(user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	var assignments []*data.Assignment

	for i := range journals {
		a, err := app.models.Assignments.GetAssignmentsByJournalID(journals[i].ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		assignments = append(assignments, a...)
	}

	sort.SliceStable(assignments, func(i, j int) bool {
		return assignments[i].Deadline.Time.After(assignments[j].Deadline.Time)
	})

	err = app.outputJSON(w, http.StatusOK, envelope{"assignments": assignments})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) setAssignmentDoneForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "sid"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	if sessionUser.ID != userID {
		app.notAllowed(w, r)
		return
	}

	if sessionUser.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "aid"))
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

	ok, err := app.models.Journals.IsUserInJournal(sessionUser.ID, assignment.JournalID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Assignments.SetAssignmentDoneForUserID(sessionUser.ID, assignment.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeAssignmentDoneForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "sid"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	if sessionUser.ID != userID {
		app.notAllowed(w, r)
		return
	}

	if sessionUser.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "aid"))
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

	ok, err := app.models.Journals.IsUserInJournal(sessionUser.ID, assignment.JournalID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Assignments.RemoveAssignmentDoneForUserID(sessionUser.ID, assignment.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
