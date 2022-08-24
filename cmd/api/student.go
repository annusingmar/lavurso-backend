package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-chi/chi/v5"
)

type StudentLatest struct {
	Marks   []*data.Mark
	Lessons []*data.Lesson
}

func (app *application) getLatestMarksLessonsForStudent(w http.ResponseWriter, r *http.Request) {
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

	var from *data.Date
	var until *data.Date

	fromDate := r.URL.Query().Get("from")
	if fromDate == "" {
		from = &data.Date{Time: helpers.ToPtr(time.Now().UTC())}
	} else {
		from, err = data.ParseDate(fromDate)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	untilDate := r.URL.Query().Get("until")
	if untilDate == "" {
		until = nil
	} else {
		until, err = data.ParseDate(untilDate)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	marks, err := app.models.Marks.GetLatestMarksForStudent(*student.ID, from, until)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	lessons, err := app.models.Lessons.GetLatestLessonsForStudent(*student.ID, from, until)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	latest := make(map[string]*StudentLatest)

	for _, m := range marks {
		if _, ok := latest[m.UpdatedAt.Format("2006-01-02")]; !ok {
			latest[m.UpdatedAt.Format("2006-01-02")] = new(StudentLatest)
		}
		latest[m.UpdatedAt.Format("2006-01-02")].Marks = append(latest[m.UpdatedAt.Format("2006-01-02")].Marks, m)
	}

	for _, l := range lessons {
		if _, ok := latest[l.Date.String()]; !ok {
			latest[l.Date.String()] = new(StudentLatest)
		}
		latest[l.Date.String()].Lessons = append(latest[l.Date.String()].Lessons, l)
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"latest": latest})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
