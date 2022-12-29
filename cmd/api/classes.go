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
)

func (app *application) listAllClasses(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var err error
	var classes []*data.ClassExt

	current := r.URL.Query().Get("current")
	if *sessionUser.Role != data.RoleAdministrator || current != "false" {
		classes, err = app.models.Classes.AllClasses(true)
	} else {
		classes, err = app.models.Classes.AllClasses(false)
	}

	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"classes": classes})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getClassesForTeacher(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	classes, err := app.models.Classes.GetCurrentYearClassesForTeacher(teacher.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"classes": classes})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createClass(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string `json:"name"`
		TeacherIDs []int  `json:"teacher_ids"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	class := &data.Class{
		Name: &input.Name,
	}

	v.Check(*class.Name != "", "name", "must be provided")

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

	err = app.models.Classes.InsertClass(class)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	if len(input.TeacherIDs) > 0 {
		err = app.models.Classes.SetClassTeachers(class.ID, input.TeacherIDs)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"class": class})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}

func (app *application) getClass(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if classID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchClass.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(classID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"class": class})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) updateClass(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if classID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchClass.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(classID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
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
		class.Name = input.Name
	}

	v := validator.NewValidator()
	v.Check(*class.Name != "", "name", "must be provided")

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

	err = app.models.Classes.UpdateClass(class)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Classes.SetClassTeachers(class.ID, input.TeacherIDs)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}

func (app *application) getStudentsInClass(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	classID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if classID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchClass.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(classID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOfClass(sessionUser.ID, class.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	users, err := app.models.Classes.GetUsersForClassID(class.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"users": users})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}
