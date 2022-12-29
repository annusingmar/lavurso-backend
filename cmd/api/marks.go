package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slices"
)

func (app *application) getMarksForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if year < 1 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, "not valid year")
		return
	}

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
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

	if sessionUser.ID != student.ID && *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOrParentOfStudent(student.ID, sessionUser.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	journals, err := app.models.Journals.GetJournalsByStudent(student.ID, year)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	marks, err := app.models.Marks.GetMarksByStudent(student.ID, year)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	type jwm struct {
		*data.NJournal
		Marks map[int][]*data.NMark `json:"marks,omitempty"`
	}

	journalsWithMarks := make([]*jwm, len(journals))
	for i, j := range journals {
		journalsWithMarks[i] = &jwm{j, make(map[int][]*data.NMark)}
	}

	for _, j := range journalsWithMarks {
		for _, m := range marks {
			if j.ID == *m.JournalID {
				if m.Course != nil {
					j.Marks[*m.Course] = append(j.Marks[*m.Course], m)
				} else {
					j.Marks[-1] = append(j.Marks[-1], m)
				}
			}
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journalsWithMarks})
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	students, err := app.models.Marks.GetStudentsMarksForLesson(lesson.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getMarksForCourse(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "jid"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	course, err := strconv.Atoi(chi.URLParam(r, "course"))
	if course < 1 || err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "invalid course")
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	students, err := app.models.Marks.GetStudentsMarksForCourse(journal.ID, course)
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getMarksForJournalSubject(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "jid"))
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	students, err := app.models.Marks.GetStudentsMarksForJournalSubject(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) setMarksForLesson(w http.ResponseWriter, r *http.Request) {
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

	var input []struct {
		StudentID int   `json:"student_id"`
		Absent    *bool `json:"absent"`
		Late      *bool `json:"late"`
		NotDone   *bool `json:"not_done"`
		Marks     []struct {
			ID      *int    `json:"id"`
			Grade   *int    `json:"grade"`
			Type    string  `json:"type"`
			Comment *string `json:"comment"`
			Remove  bool    `json:"remove"`
		} `json:"marks"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	currentTime := time.Now().UTC()
	v := validator.NewValidator()

	allMarkIDs, err := app.models.Marks.GetMarkIDsForLesson(lesson.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	allStudentIDs, err := app.models.Journals.GetStudentIDsForJournal(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	allGradeIDs, err := app.models.Grades.GetAllGradeIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	newMark := func(index int, ID int, studentID int, mtype string, comment *string, grade *int) *model.Marks {
		switch mtype {
		case data.MarkCommonGrade:
			if grade == nil {
				v.Add("grade", fmt.Sprintf("%d: must be provided", index))
				return nil
			} else if !slices.Contains(allGradeIDs, *grade) {
				v.Add("grade", fmt.Sprintf("%d: invalid grade ID", index))
				return nil
			}
		case data.MarkNoticeBad, data.MarkNoticeNeutral, data.MarkNoticeGood:
			if comment == nil || *comment == "" {
				v.Add("comment", fmt.Sprintf("%d: must be provided and not empty", index))
				return nil
			}
		case data.MarkAbsent, data.MarkLate, data.MarkNotDone:
		default:
			panic("invalid or no type: " + mtype)
		}

		if comment != nil && *comment == "" {
			comment = nil
		}
		if grade != nil && *grade == 0 {
			grade = nil
		}

		if mtype == data.MarkCommonGrade {
			mtype = data.MarkLessonGrade
		}

		return &model.Marks{
			ID:        ID,
			UserID:    &studentID,
			LessonID:  &lesson.ID,
			Course:    lesson.Course,
			JournalID: &journal.ID,
			TeacherID: &sessionUser.ID,
			Type:      &mtype,
			GradeID:   grade,
			Comment:   comment,
			CreatedAt: &currentTime,
			UpdatedAt: &currentTime,
		}
	}

	var insertMarks []*model.Marks
	var updateMarks []*model.Marks
	var deletedMarkIDs []int

	var deletedMarksByStudentIDType []data.MarkByStudentIDType

student:
	for _, s := range input {
		if s.StudentID < 1 {
			continue
		}
		if !slices.Contains(allStudentIDs, s.StudentID) {
			v.Add("student_id", fmt.Sprintf("%s: %d", data.ErrUserNotInJournal.Error(), s.StudentID))
			continue student
		}

		if s.Absent != nil {
			if *s.Absent {
				insertMarks = append(insertMarks, newMark(0, 0, s.StudentID, data.MarkAbsent, nil, nil))
			} else {
				deletedMarksByStudentIDType = append(deletedMarksByStudentIDType, data.MarkByStudentIDType{StudentID: s.StudentID, Type: data.MarkAbsent})
			}
		}

		if s.Late != nil {
			if *s.Late {
				insertMarks = append(insertMarks, newMark(0, 0, s.StudentID, data.MarkLate, nil, nil))
			} else {
				deletedMarksByStudentIDType = append(deletedMarksByStudentIDType, data.MarkByStudentIDType{StudentID: s.StudentID, Type: data.MarkLate})
			}
		}

		if s.NotDone != nil {
			if *s.NotDone {
				insertMarks = append(insertMarks, newMark(0, 0, s.StudentID, data.MarkNotDone, nil, nil))
			} else {
				deletedMarksByStudentIDType = append(deletedMarksByStudentIDType, data.MarkByStudentIDType{StudentID: s.StudentID, Type: data.MarkNotDone})
			}
		}

	marks:
		for mi, m := range s.Marks {
			if m.Type != data.MarkCommonGrade && m.Type != data.MarkNoticeBad &&
				m.Type != data.MarkNoticeNeutral && m.Type != data.MarkNoticeGood && !m.Remove {
				v.Add("type", fmt.Sprintf("%s: %d", data.ErrNoSuchType.Error(), mi))
				continue marks
			}
			if m.ID != nil {
				if !slices.Contains(allMarkIDs, *m.ID) {
					v.Add("mark_id", fmt.Sprintf("%s: %d at %d", data.ErrNoSuchMark.Error(), *m.ID, mi))
					continue marks
				}
				if m.Remove {
					deletedMarkIDs = append(deletedMarkIDs, *m.ID)
				} else {
					updateMarks = append(updateMarks, newMark(mi, *m.ID, 0, m.Type, m.Comment, m.Grade))
				}
			} else {
				if m.Remove {
					continue marks
				}
				insertMarks = append(insertMarks, newMark(mi, 0, s.StudentID, m.Type, m.Comment, m.Grade))
			}
		}
	}

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	tx, err := app.models.Marks.DB.Begin()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	defer tx.Rollback()

	if len(insertMarks) > 0 {
		err := app.models.Marks.InsertMarks(tx, insertMarks)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(updateMarks) > 0 {
		err := app.models.Marks.UpdateMarks(tx, updateMarks)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(deletedMarkIDs) > 0 {
		err := app.models.Marks.DeleteMarks(tx, deletedMarkIDs)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(deletedMarksByStudentIDType) > 0 {
		err := app.models.Marks.DeleteMarksByStudentIDType(tx, deletedMarksByStudentIDType)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) setMarksForCourse(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "jid"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	course, err := strconv.Atoi(chi.URLParam(r, "course"))
	if course < 1 || err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "invalid course")
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input []struct {
		StudentID int `json:"student_id"`
		Marks     []struct {
			ID      *int    `json:"id"`
			Grade   int     `json:"grade"`
			Comment *string `json:"comment"`
			Remove  bool    `json:"remove"`
		} `json:"marks"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	currentTime := time.Now().UTC()
	v := validator.NewValidator()

	allMarkIDs, err := app.models.Marks.GetMarkIDsForCourse(journal.ID, course)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	allStudentIDs, err := app.models.Journals.GetStudentIDsForJournal(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	allGradeIDs, err := app.models.Grades.GetAllGradeIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	newMark := func(index int, ID int, studentID int, comment *string, grade int) *model.Marks {
		if grade == 0 {
			v.Add("grade", fmt.Sprintf("%d: must be provided", index))
			return nil
		} else if !slices.Contains(allGradeIDs, grade) {
			v.Add("grade", fmt.Sprintf("%d: invalid grade ID", index))
			return nil
		}

		if comment != nil && *comment == "" {
			comment = nil
		}

		return &model.Marks{
			ID:        ID,
			UserID:    &studentID,
			Course:    &course,
			JournalID: &journal.ID,
			TeacherID: &sessionUser.ID,
			Type:      helpers.ToPtr(data.MarkCourseGrade),
			GradeID:   &grade,
			Comment:   comment,
			CreatedAt: &currentTime,
			UpdatedAt: &currentTime,
		}
	}

	var insertMarks []*model.Marks
	var updateMarks []*model.Marks
	var deletedMarkIDs []int

student:
	for _, s := range input {
		if s.StudentID < 1 {
			continue
		}

		if !slices.Contains(allStudentIDs, s.StudentID) {
			v.Add("student_id", fmt.Sprintf("%s: %d", data.ErrUserNotInJournal.Error(), s.StudentID))
			continue student
		}
	marks:
		for mi, m := range s.Marks {
			if m.ID != nil {
				if !slices.Contains(allMarkIDs, *m.ID) {
					v.Add("mark_id", fmt.Sprintf("%s: %d at %d", data.ErrNoSuchMark.Error(), *m.ID, mi))
					continue marks
				}
				if m.Remove {
					deletedMarkIDs = append(deletedMarkIDs, *m.ID)
				} else {
					updateMarks = append(updateMarks, newMark(mi, *m.ID, 0, m.Comment, m.Grade))
				}
			} else {
				if m.Remove {
					continue marks
				}
				insertMarks = append(insertMarks, newMark(mi, 0, s.StudentID, m.Comment, m.Grade))
			}
		}
	}

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	tx, err := app.models.Marks.DB.Begin()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	defer tx.Rollback()

	if len(insertMarks) > 0 {
		err := app.models.Marks.InsertMarks(tx, insertMarks)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(updateMarks) > 0 {
		err := app.models.Marks.UpdateMarks(tx, updateMarks)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(deletedMarkIDs) > 0 {
		err := app.models.Marks.DeleteMarks(tx, deletedMarkIDs)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) setMarksForJournalSubject(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "jid"))
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input []struct {
		StudentID int `json:"student_id"`
		Marks     []struct {
			ID      *int    `json:"id"`
			Grade   int     `json:"grade"`
			Comment *string `json:"comment"`
			Remove  bool    `json:"remove"`
		} `json:"marks"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	currentTime := time.Now().UTC()
	v := validator.NewValidator()

	allMarkIDs, err := app.models.Marks.GetMarkIDsForJournalSubject(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	allStudentIDs, err := app.models.Journals.GetStudentIDsForJournal(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	allGradeIDs, err := app.models.Grades.GetAllGradeIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	newMark := func(index int, ID int, studentID int, comment *string, grade int) *model.Marks {
		if grade == 0 {
			v.Add("grade", fmt.Sprintf("%d: must be provided", index))
			return nil
		} else if !slices.Contains(allGradeIDs, grade) {
			v.Add("grade", fmt.Sprintf("%d: invalid grade ID", index))
			return nil
		}

		if comment != nil && *comment == "" {
			comment = nil
		}

		return &model.Marks{
			ID:        ID,
			UserID:    &studentID,
			JournalID: &journal.ID,
			TeacherID: &sessionUser.ID,
			Type:      helpers.ToPtr(data.MarkSubjectGrade),
			GradeID:   &grade,
			Comment:   comment,
			CreatedAt: &currentTime,
			UpdatedAt: &currentTime,
		}
	}

	var insertMarks []*model.Marks
	var updateMarks []*model.Marks
	var deletedMarkIDs []int

student:
	for _, s := range input {
		if s.StudentID < 1 {
			continue
		}

		if !slices.Contains(allStudentIDs, s.StudentID) {
			v.Add("student_id", fmt.Sprintf("%s: %d", data.ErrUserNotInJournal.Error(), s.StudentID))
			continue student
		}
	marks:
		for mi, m := range s.Marks {
			if m.ID != nil {
				if !slices.Contains(allMarkIDs, *m.ID) {
					v.Add("mark_id", fmt.Sprintf("%s: %d at %d", data.ErrNoSuchMark.Error(), *m.ID, mi))
					continue marks
				}
				if m.Remove {
					deletedMarkIDs = append(deletedMarkIDs, *m.ID)
				} else {
					updateMarks = append(updateMarks, newMark(mi, *m.ID, 0, m.Comment, m.Grade))
				}
			} else {
				if m.Remove {
					continue marks
				}
				insertMarks = append(insertMarks, newMark(mi, 0, s.StudentID, m.Comment, m.Grade))
			}
		}
	}

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	tx, err := app.models.Marks.DB.Begin()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	defer tx.Rollback()

	if len(insertMarks) > 0 {
		err := app.models.Marks.InsertMarks(tx, insertMarks)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(updateMarks) > 0 {
		err := app.models.Marks.UpdateMarks(tx, updateMarks)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if len(deletedMarkIDs) > 0 {
		err := app.models.Marks.DeleteMarks(tx, deletedMarkIDs)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getLessonsForStudentsJournalsCourse(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "sid"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
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

	if sessionUser.ID != student.ID && *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOrParentOfStudent(student.ID, sessionUser.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	journalID, err := strconv.Atoi(chi.URLParam(r, "jid"))
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

	ok, err := app.models.Journals.IsUserInJournal(student.ID, journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.writeErrorResponse(w, r, http.StatusTeapot, "user not in journal")
		return
	}

	course, err := strconv.Atoi(r.URL.Query().Get("course"))
	if err != nil || course < 1 {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "invalid course")
		return
	}

	lessons, err := app.models.Lessons.GetLessonsAndStudentMarksByJournalID(student.ID, journal.ID, course)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"lessons": lessons})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
