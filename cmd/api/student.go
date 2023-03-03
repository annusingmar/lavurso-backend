package main

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-chi/chi/v5"
)

type StudentLatestByDate struct {
	Date    string            `json:"date"`
	Marks   []*data.MarkExt   `json:"marks"`
	Lessons []*data.LessonExt `json:"lessons"`
}

func (app *application) getLatestMarksLessonsForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	var from *types.Date
	var until *types.Date

	fromDate := r.URL.Query().Get("from")
	if fromDate == "" {
		from = &types.Date{Time: helpers.ToPtr(time.Now().UTC())}
	} else {
		from, err = types.ParseDate(fromDate)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	untilDate := r.URL.Query().Get("until")
	if untilDate == "" {
		until = nil
	} else {
		until, err = types.ParseDate(untilDate)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	marks, err := app.models.Marks.GetLatestMarksForStudent(student.ID, from, until)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	lessons, err := app.models.Lessons.GetLatestLessonsForStudent(student.ID, from, until)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	var latest []*StudentLatestByDate
	dateIndexMap := make(map[string]int)

	for _, m := range marks {
		dateString := m.UpdatedAt.Format("2006-01-02")
		if val, ok := dateIndexMap[dateString]; !ok {
			latest = append(latest, &StudentLatestByDate{Date: dateString, Marks: []*data.MarkExt{m}})
			dateIndexMap[dateString] = len(latest) - 1
		} else {
			latest[val].Marks = append(latest[val].Marks, m)
		}
	}

	for _, l := range lessons {
		dateString := l.Date.Format("2006-01-02")
		if val, ok := dateIndexMap[dateString]; !ok {
			latest = append(latest, &StudentLatestByDate{Date: dateString, Lessons: []*data.LessonExt{l}})
			dateIndexMap[dateString] = len(latest) - 1
		} else {
			latest[val].Lessons = append(latest[val].Lessons, l)
		}
	}

	sort.Slice(latest, func(i, j int) bool {
		t1, _ := time.Parse("2006-01-02", latest[i].Date)
		t2, _ := time.Parse("2006-01-02", latest[j].Date)
		return t1.After(t2)
	})

	err = app.outputJSON(w, http.StatusOK, envelope{"latest": latest})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
