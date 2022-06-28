package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createLesson(w http.ResponseWriter, r *http.Request) {
	var input struct {
		JournalID   int       `json:"journal_id"`
		Description string    `json:"description"`
		Date        data.Date `json:"date"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	lesson := &data.Lesson{
		JournalID:   input.JournalID,
		Description: input.Description,
		Date:        input.Date,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Version:     1,
	}

	journal, err := app.models.Journals.GetJournalByID(lesson.JournalID)
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

	err = app.models.Lessons.InsertLesson(lesson)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"lesson": lesson})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) getLesson(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	lessonID, err := strconv.Atoi(params.ByName("id"))
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

	err = app.outputJSON(w, http.StatusOK, envelope{"lesson": lesson})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) updateLesson(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	lessonID, err := strconv.Atoi(params.ByName("id"))
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

	var input struct {
		Description *string    `json:"description"`
		Date        *data.Date `json:"date"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Description != nil {
		lesson.Description = *input.Description
	}
	if input.Date != nil {
		lesson.Date = *input.Date
	}

	lesson.UpdatedAt = time.Now().UTC()

	err = app.models.Lessons.UpdateLesson(lesson)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"lesson": lesson})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
