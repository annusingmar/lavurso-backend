package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) createLesson(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		JournalID   int        `json:"journal_id"`
		Description string     `json:"description"`
		Date        types.Date `json:"date"`
		Course      int        `json:"course"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	time := time.Now().UTC()

	lesson := &data.NLesson{
		Lessons: model.Lessons{
			JournalID:   &input.JournalID,
			Description: &input.Description,
			Date:        &input.Date,
			Course:      &input.Course,
			CreatedAt:   &time,
			UpdatedAt:   &time,
		},
	}

	if lesson.Date.Time.IsZero() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, types.ErrInvalidDateFormat.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(*lesson.Course > 0, "course", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	journal, err := app.models.Journals.GetJournalByID(lesson.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Lessons.InsertLesson(lesson)
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

func (app *application) getLesson(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	lessonID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if lessonID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchLesson.Error())
		return
	}

	lesson, err := app.models.Lessons.GetLessonByID(lessonID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchLesson):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	journal, err := app.models.Journals.GetJournalByID(lesson.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"lesson": lesson})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) updateLesson(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	lessonID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if lessonID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchLesson.Error())
		return
	}

	lesson, err := app.models.Lessons.GetLessonByID(lessonID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchLesson):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	journal, err := app.models.Journals.GetJournalByID(lesson.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		Description *string     `json:"description"`
		Date        *types.Date `json:"date"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Description != nil {
		lesson.Description = input.Description
	}
	if input.Date != nil {
		lesson.Date = input.Date
	}

	lesson.UpdatedAt = helpers.ToPtr(time.Now().UTC())

	err = app.models.Lessons.UpdateLesson(lesson)
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

func (app *application) deleteLesson(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	lessonID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if lessonID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchLesson.Error())
		return
	}

	lesson, err := app.models.Lessons.GetLessonByID(lessonID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchLesson):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	journal, err := app.models.Journals.GetJournalByID(lesson.Journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Lessons.DeleteLesson(lesson.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getLessonsForJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	course, err := strconv.Atoi(r.URL.Query().Get("course"))
	if course < 0 || err != nil {
		course = 1
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	lessons, err := app.models.Lessons.GetLessonsByJournalID(journal.ID, course)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"lessons": lessons})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
