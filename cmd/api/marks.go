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

func (app *application) addMark(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		UserID    int     `json:"user_id"`
		LessonID  *int    `json:"lesson_id"`
		Course    *int    `json:"course"`
		JournalID *int    `json:"journal_id"`
		GradeID   *int    `json:"grade_id"`
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

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	mark := new(data.Mark)
	mark.Grade = new(data.Grade)
	mark.By = new(data.User)
	mark.Lesson = &data.Lesson{Date: new(data.Date)}

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

			mark.Grade.ID = grade.ID
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
		mark.Lesson.ID = lesson.ID

		mark.Course = lesson.Course
		mark.JournalID = &lesson.Journal.ID
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
		mark.Grade.ID = grade.ID

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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
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

	mark.Comment = input.Comment
	mark.By.ID = sessionUser.ID
	mark.At = time.Now().UTC()

	err = app.models.Marks.InsertMark(mark)
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

func (app *application) deleteMark(w http.ResponseWriter, r *http.Request) {
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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	err = app.models.Marks.DeleteMark(mark.ID)
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

func (app *application) updateMark(w http.ResponseWriter, r *http.Request) {
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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	var input struct {
		GradeID *int    `json:"grade_id"`
		Comment *string `json:"comment"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var updated bool

	oldMark := *mark
	oldMark.MarkID = &oldMark.ID

	switch mark.Type {
	case data.MarkLessonGrade, data.MarkCourseGrade, data.MarkSubjectGrade:
		if input.GradeID != nil {
			v := validator.NewValidator()

			v.Check(*input.GradeID > 0, "grade_id", "must be provided and valid")

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

			if *mark.Grade.ID != *grade.ID {
				updated = true
				mark.Grade.ID = grade.ID
			}
		}
	}

	if input.Comment != nil {
		if mark.Comment == nil || *mark.Comment != *input.Comment {
			updated = true
			mark.Comment = input.Comment
		}
	}

	if updated {
		mark.At = time.Now().UTC()
		mark.By.ID = sessionUser.ID

		err = app.models.Marks.UpdateMark(mark)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		err = app.models.Marks.InsertOldMark(&oldMark)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		err = app.models.Journals.SetJournalLastUpdated(journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) getMark(w http.ResponseWriter, r *http.Request) {
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

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		journal, err := app.models.Journals.GetJournalByID(*mark.JournalID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if sessionUser.ID != journal.Teacher.ID {
			app.notAllowed(w, r)
			return
		}
	case data.RoleParent:
		ok, err := app.models.Users.IsParentOfChild(sessionUser.ID, mark.UserID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		if sessionUser.ID != mark.UserID {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"mark": mark})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getMarksForStudent(w http.ResponseWriter, r *http.Request) {
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

	marks, err := app.models.Marks.GetMarksByStudent(user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"marks": marks})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getGradesForJournal(w http.ResponseWriter, r *http.Request) {
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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	gradeType := r.URL.Query().Get("grade_type")
	if gradeType == "" {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "missing grade type")
		return
	}

	students, err := app.models.Journals.GetUsersByJournalID(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	var marks []*data.Mark

	switch gradeType {
	case data.MarkSubjectGrade:
		marks, err = app.models.Marks.GetSubjectGradesByJournalID(journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	case data.MarkCourseGrade:
		course, err := strconv.Atoi(r.URL.Query().Get("course"))
		if err != nil || course < 1 {
			marks, err = app.models.Marks.GetAllCoursesGradesByJournalID(journal.ID)
		} else {
			marks, err = app.models.Marks.GetCourseGradesByJournalID(journal.ID, course)
		}
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	case data.MarkLessonGrade:
		course, err := strconv.Atoi(r.URL.Query().Get("course"))
		if err != nil || course < 1 {
			app.writeErrorResponse(w, r, http.StatusBadRequest, "invalid course")
			return
		}
		marks, err = app.models.Marks.GetLessonGradesByCourseAndJournalID(journal.ID, course)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	default:
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNoSuchType)
		return
	}

	for _, mark := range marks {
		for _, student := range students {
			if student.ID == mark.UserID {
				student.Marks = append(student.Marks, mark)
			}
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getMarksForLesson(w http.ResponseWriter, r *http.Request) {
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

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	students, err := app.models.Journals.GetUsersByJournalID(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	marks, err := app.models.Marks.GetMarksByLessonID(*lesson.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	for _, mark := range marks {
		for _, student := range students {
			if student.ID == mark.UserID {
				student.Marks = append(student.Marks, mark)
			}
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

// func (app *application) getMarksForStudentsJournal(w http.ResponseWriter, r *http.Request) {
// 	sessionUser := app.getUserFromContext(r)

// 	userID, err := strconv.Atoi(chi.URLParam(r, "sid"))
// 	if userID < 0 || err != nil {
// 		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
// 		return
// 	}

// 	switch sessionUser.Role {
// 	case data.RoleAdministrator:
// 	case data.RoleParent:
// 		ok, err := app.models.Users.IsParentOfChild(sessionUser.ID, userID)
// 		if err != nil {
// 			app.writeInternalServerError(w, r, err)
// 			return
// 		}
// 		if !ok {
// 			app.notAllowed(w, r)
// 			return
// 		}
// 	case data.RoleTeacher:
// 	case data.RoleStudent:
// 		if userID != sessionUser.ID {
// 			app.notAllowed(w, r)
// 			return
// 		}
// 	default:
// 		app.notAllowed(w, r)
// 		return
// 	}

// 	user, err := app.models.Users.GetUserByID(userID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrNoSuchUser):
// 			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
// 		default:
// 			app.writeInternalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	if user.Role != data.RoleStudent {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
// 		return
// 	}

// 	journalID, err := strconv.Atoi(chi.URLParam(r, "jid"))
// 	if journalID < 0 || err != nil {
// 		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
// 		return
// 	}

// 	journal, err := app.models.Journals.GetJournalByID(journalID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrNoSuchJournal):
// 			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
// 		default:
// 			app.writeInternalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	if sessionUser.Role == data.RoleTeacher && journal.Teacher.ID != sessionUser.ID {
// 		app.notAllowed(w, r)
// 		return
// 	}

// 	ok, err := app.models.Journals.IsUserInJournal(user.ID, journal.ID)
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 		return
// 	}
// 	if !ok {
// 		app.writeErrorResponse(w, r, http.StatusTeapot, "user not in journal")
// 		return
// 	}

// 	marks, err := app.models.Marks.GetMarksByUserIDAndJournalID(user.ID, journal.ID)
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 		return
// 	}

// 	err = app.outputJSON(w, http.StatusOK, envelope{"marks": marks})
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 	}
// }

func (app *application) getPreviousMarksForMark(w http.ResponseWriter, r *http.Request) {
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

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		journal, err := app.models.Journals.GetJournalByID(*mark.JournalID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if sessionUser.ID != journal.Teacher.ID {
			app.notAllowed(w, r)
			return
		}
	case data.RoleParent:
		ok, err := app.models.Users.IsParentOfChild(sessionUser.ID, mark.UserID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		if sessionUser.ID != mark.UserID {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	previousMarks, err := app.models.Marks.GetOldMarksByMarkID(mark.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"marks": previousMarks})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
