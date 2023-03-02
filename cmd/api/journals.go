package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slices"
)

func (app *application) listAllJournals(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if year < 1 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, "not valid year")
		return
	}

	journals, err := app.models.Journals.AllJournals(year)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournal(w http.ResponseWriter, r *http.Request) {
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journal": journal})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		Name      string `json:"name"`
		SubjectID int    `json:"subject_id"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	journal := &data.Journal{
		Name:      &input.Name,
		SubjectID: &input.SubjectID,
	}

	v.Check(*journal.Name != "", "name", "must be provided")
	v.Check(*journal.SubjectID > 0, "subject_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	_, err = app.models.Subjects.GetSubjectByID(*journal.SubjectID, false)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSubject):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	year, err := app.models.Years.GetCurrentYear()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	} else if year == nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNoCurrentYear.Error())
		return
	}

	journal.YearID = &year.ID

	err = app.models.Journals.InsertJournal(journal, sessionUser.ID)
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

func (app *application) updateJournal(w http.ResponseWriter, r *http.Request) {
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		Name       *string `json:"name"`
		TeacherIDs []int   `json:"teacher_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name != nil {
		journal.Name = input.Name
	}

	v := validator.NewValidator()

	v.Check(*journal.Name != "", "name", "must be provided")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	allUserIDs, err := app.models.Users.GetAllUserIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	badIDs := helpers.VerifyExistsInSlice(input.TeacherIDs, allUserIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchUsers.Error(), badIDs))
		return
	}

	if !slices.Contains(input.TeacherIDs, sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		input.TeacherIDs = append(input.TeacherIDs, sessionUser.ID)
	}

	err = app.models.Journals.UpdateJournal(journal, input.TeacherIDs)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) deleteJournal(w http.ResponseWriter, r *http.Request) {
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

	err = app.models.Journals.DeleteJournal(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournalsForTeacher(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if year < 1 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, "not valid year")
		return
	}

	teacherID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if teacherID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	if teacherID != sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	teacher, err := app.models.Users.GetUserByID(teacherID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *teacher.Role != data.RoleAdministrator && *teacher.Role != data.RoleTeacher {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	journals, err := app.models.Journals.GetJournalsForTeacher(teacher.ID, year)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addStudentsToJournal(w http.ResponseWriter, r *http.Request) {
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		StudentIDs []int `json:"student_ids"`
		ClassIDs   []int `json:"class_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	allStudentIDs, err := app.models.Users.GetAllStudentIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs := helpers.VerifyExistsInSlice(input.StudentIDs, allStudentIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchStudents.Error(), badIDs))
		return
	}

	allClassIDs, err := app.models.Classes.GetAllClassIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs = helpers.VerifyExistsInSlice(input.ClassIDs, allClassIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchClass.Error(), badIDs))
		return
	}

	if len(input.StudentIDs) > 0 {
		err = app.models.Journals.InsertStudentsIntoJournal(input.StudentIDs, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.ClassIDs {
		users, err := app.models.Classes.GetUsersForClassID(id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		var ids []int
		for _, u := range users {
			ids = append(ids, u.ID)
		}

		if len(ids) > 0 {
			err = app.models.Journals.InsertStudentsIntoJournal(ids, journal.ID)
			if err != nil {
				app.writeInternalServerError(w, r, err)
				return
			}
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeStudentFromJournal(w http.ResponseWriter, r *http.Request) {
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		StudentID int `json:"student_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.StudentID > 0, "student_id", "must be provided and valid")
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	user, err := app.models.Users.GetUserByID(input.StudentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	err = app.models.Journals.DeleteStudentFromJournal(user.ID, journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserNotInJournal):
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getStudentsForJournal(w http.ResponseWriter, r *http.Request) {
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

	if !journal.IsUserTeacherOfJournal(sessionUser.ID) && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	students, err := app.models.Journals.GetStudentsByJournalID(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
