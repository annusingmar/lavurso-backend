package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
)

func (app *application) addMark(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID    int     `json:"user_id"`
		LessonID  *int    `json:"lesson_id"`
		Course    *int    `json:"course"`
		JournalID *int    `json:"journal_id"`
		GradeID   *int    `json:"grade_id"`
		SubjectID *int    `json:"subject_id"`
		Comment   *string `json:"comment"`
		Type      string  `json:"type"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.UserID > 0, "user_id", "must be provided and valid")
	v.Check(input.Type != "", "type", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	// student check
	user, err := app.models.Users.GetUserByID(input.UserID)
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

	mark := &data.Mark{}

	mark.UserID = user.ID
	mark.Type = input.Type

	switch mark.Type {
	case data.MarkLessonGrade, data.MarkNotDone, data.MarkNoticeGood, data.MarkNoticeNeutral, data.MarkNoticeBad, data.MarkAbsent, data.MarkLate:
		v.Check(input.LessonID != nil && *input.LessonID > 0, "lesson_id", "must be provided and valid")

		if !v.Valid() {
			app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
			return
		}

		if input.Type == data.MarkLessonGrade {
			v.Check(input.GradeID != nil && *input.GradeID > 0, "grade_id", "must be provided and valid")

			if !v.Valid() {
				app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
				return
			}

			grade, err := app.models.Grades.GetGradeByID(*input.GradeID)
			if err != nil {
				switch {
				case errors.Is(err, data.ErrNoSuchGrade):
					app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
				default:
					app.writeInternalServerError(w, r, err)
				}
				return
			}

			mark.GradeID = &grade.ID
		}

		// lesson check
		lesson, err := app.models.Lessons.GetLessonByID(*input.LessonID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoSuchLesson):
				app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}
		mark.LessonID = &lesson.ID

		mark.Course = &lesson.Course
		mark.JournalID = &lesson.JournalID
	case data.MarkCourseGrade, data.MarkSubjectGrade:
		if mark.Type == data.MarkCourseGrade {
			v.Check(input.Course != nil && *input.Course > 0, "course", "must be provided and valid")
			mark.Course = input.Course
		}
		v.Check(input.JournalID != nil && *input.JournalID > 0, "journal_id", "must be provided and valid")
		v.Check(input.GradeID != nil && *input.GradeID > 0, "grade_id", "must be provided and valid")

		if !v.Valid() {
			app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
			return
		}

		grade, err := app.models.Grades.GetGradeByID(*input.GradeID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoSuchGrade):
				app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}
		mark.GradeID = &grade.ID

		mark.JournalID = input.JournalID
	default:
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchType.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(*mark.JournalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}
	mark.JournalID = &journal.ID

	ok, err := app.models.Journals.IsUserInJournal(mark.UserID, *mark.JournalID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.writeErrorResponse(w, r, http.StatusTeapot, "user not in journal")
		return
	}

	subject, err := app.models.Subjects.GetSubjectByID(journal.SubjectID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSubject):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}
	mark.SubjectID = &subject.ID

	mark.Comment = input.Comment
	mark.Current = true
	mark.By = 1 // temporary
	mark.At = time.Now().UTC()

	err = app.models.Marks.InsertMark(mark)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"mark": mark})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
