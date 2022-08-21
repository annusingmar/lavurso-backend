package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
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

	journal, err := app.models.Journals.GetJournalByID(assignment.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *journal.Teacher.ID != *sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
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

	v := validator.NewValidator()
	v.Check(input.JournalID > 0, "journal_id", "must be provided and valid")
	v.Check(input.Type == data.AssignmentHomework || input.Type == data.AssignmentTest, "type", "must be provided and valid")
	v.Check(input.Deadline.After(time.Now().UTC()), "deadline", "must not be in the past")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	assignment := &data.Assignment{
		Journal:     &data.Journal{ID: input.JournalID},
		Description: input.Description,
		Deadline:    input.Deadline,
		Type:        input.Type,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if assignment.Deadline.Time.IsZero() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrInvalidDateFormat.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(assignment.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *journal.Teacher.ID != *sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
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

	journal, err := app.models.Journals.GetJournalByID(assignment.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *journal.Teacher.ID != *sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	if assignment.Deadline.Before(time.Now().UTC()) {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "can't edit past assignment")
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
	v.Check(assignment.Deadline.After(time.Now().UTC()), "deadline", "must not be in the past")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	assignment.UpdatedAt = time.Now().UTC()

	err = app.models.Assignments.UpdateAssignment(assignment)
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

	journal, err := app.models.Journals.GetJournalByID(assignment.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *journal.Teacher.ID != *sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	if assignment.Deadline.Before(time.Now().UTC()) {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "can't edit past assignment")
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

	if *journal.Teacher.ID != *sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
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

	switch *sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleParent:
		ok, err := app.models.Users.IsParentOfChild(*sessionUser.ID, userID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		if *sessionUser.ID != userID {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	student, err := app.models.Users.GetStudentByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var startDate *data.Date
	var endDate *data.Date

	sd := r.URL.Query().Get("start_date")
	if sd == "" {
		startDate = &data.Date{Time: helpers.ToPtr(time.Now().UTC())}
	} else {
		startDate, err = data.ParseDate(sd)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	ed := r.URL.Query().Get("end_date")
	if ed == "" {
		endDate = &data.Date{Time: helpers.ToPtr(time.Now().UTC().AddDate(0, 0, 7))}
	} else {
		endDate, err = data.ParseDate(ed)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
		if endDate.Before(*startDate.Time) {
			app.writeErrorResponse(w, r, http.StatusBadRequest, "end date is before start date")
			return
		}
	}

	assignments, err := app.models.Assignments.GetAssignmentsForStudent(*student.ID, startDate, endDate)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

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

	if *sessionUser.ID != userID {
		app.notAllowed(w, r)
		return
	}

	if *sessionUser.Role != data.RoleStudent {
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

	ok, err := app.models.Journals.IsUserInJournal(*sessionUser.ID, assignment.Journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Assignments.SetAssignmentDoneForUserID(*sessionUser.ID, assignment.ID)
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

	if *sessionUser.ID != userID {
		app.notAllowed(w, r)
		return
	}

	if *sessionUser.Role != data.RoleStudent {
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

	ok, err := app.models.Journals.IsUserInJournal(*sessionUser.ID, assignment.Journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Assignments.RemoveAssignmentDoneForUserID(*sessionUser.ID, assignment.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
