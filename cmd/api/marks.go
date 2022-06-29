package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

func (app *application) addLessonGrade(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID   int    `json:"user_id"`
		LessonID int    `json:"lesson_id"`
		GradeID  int    `json:"grade_id"`
		Comment  string `json:"comment"`
		By       int    `json:"by"` // will be removed later with auth middleware
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	mark := &data.Mark{
		UserID:   input.UserID,
		LessonID: &input.LessonID,
		GradeID:  &input.GradeID,
		Comment:  &input.Comment,
		Type:     data.MarkLessonGrade,
		Current:  true,
		By:       input.By,
		At:       time.Now().UTC(),
	}

	// student check
	user, err := app.models.Users.GetUserByID(mark.UserID)
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

	// lesson check
	_, err = app.models.Lessons.GetLessonByID(*mark.LessonID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchLesson):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	// grade check

	app.models.Marks.InsertMark(mark)

	_ = app.outputJSON(w, http.StatusOK, envelope{"mark": mark})
}
