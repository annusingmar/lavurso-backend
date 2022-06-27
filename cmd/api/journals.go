package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) listAllJournals(w http.ResponseWriter, r *http.Request) {
	journals, err := app.models.Journals.AllJournals()
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
	params := httprouter.ParamsFromContext(r.Context())
	journalID, err := strconv.Atoi(params.ByName("id"))
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

	err = app.outputJSON(w, http.StatusOK, envelope{"journal": journal})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createJournal(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string `json:"name"`
		TeacherID int    `json:"teacher_id"`
		SubjectID int    `json:"subject_id"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	journal := &data.Journal{
		Name:      input.Name,
		TeacherID: input.TeacherID,
		SubjectID: input.SubjectID,
		Archived:  false,
	}

	v.Check(journal.Name != "", "name", "must be provided")
	v.Check(journal.TeacherID > 0, "teacher_id", "must be provided and valid")
	v.Check(journal.SubjectID > 0, "subject_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	teacher, err := app.models.Users.GetUserByID(journal.TeacherID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if teacher.Role != data.Administrator {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	_, err = app.models.Subjects.GetSubjectByID(journal.SubjectID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSubject):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Journals.InsertJournal(journal)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"journal": journal})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}

func (app *application) updateJournal(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	journalID, err := strconv.Atoi(params.ByName("id"))
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

	var input struct {
		Name      *string `json:"name"`
		TeacherID *int    `json:"teacher_id"`
		SubjectID *int    `json:"subject_id"`
		Archived  *bool   `json:"archived"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name != nil {
		journal.Name = *input.Name
	}
	if input.TeacherID != nil {
		journal.TeacherID = *input.TeacherID
	}
	if input.SubjectID != nil {
		journal.SubjectID = *input.SubjectID
	}
	if input.Archived != nil {
		journal.Archived = *input.Archived
	}

	v := validator.NewValidator()

	v.Check(journal.Name != "", "name", "must be provided")
	v.Check(journal.TeacherID > 0, "teacher_id", "must be provided and valid")
	v.Check(journal.SubjectID > 0, "subject_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	teacher, err := app.models.Users.GetUserByID(journal.TeacherID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if teacher.Role != data.Administrator {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	_, err = app.models.Subjects.GetSubjectByID(journal.SubjectID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSubject):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Journals.UpdateJournal(journal)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journal": journal})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournalsForTeacher(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	teacherID, err := strconv.Atoi(params.ByName("id"))
	if teacherID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
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

	if teacher.Role != data.Administrator {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	journals, err := app.models.Journals.GetJournalsForTeacher(teacher.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addStudentToJournal(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	userID, err := strconv.Atoi(params.ByName("id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
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

	if user.Role != data.Student {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	var input struct {
		JournalID int `json:"journal_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.JournalID > 0, "journal_id", "must be provided and valid")
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	journal, err := app.models.Journals.GetJournalByID(input.JournalID)
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

	err = app.models.Journals.InsertUserIntoJournal(user.ID, journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserAlreadyInJournal):
			app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
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
	params := httprouter.ParamsFromContext(r.Context())
	journalID, err := strconv.Atoi(params.ByName("id"))
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

	students, err := app.models.Journals.GetUsersByJournalID(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournalsForStudent(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	userID, err := strconv.Atoi(params.ByName("id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
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

	if user.Role != data.Student {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	journals, err := app.models.Journals.GetJournalsByUserID(user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}
