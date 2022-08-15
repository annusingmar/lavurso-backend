package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) listAllClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := app.models.Classes.AllClasses()
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

	if teacherID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	// var archived bool

	// archivedParam := r.URL.Query().Get("archived")

	// if archivedParam == "true" {
	// 	archived = true
	// } else {
	// 	archived = false
	// }

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

	if teacher.Role != data.RoleAdministrator && teacher.Role != data.RoleTeacher {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	classes, err := app.models.Classes.GetClassesForTeacher(teacher.ID)
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
		Name      string `json:"name"`
		TeacherID int    `json:"teacher_id"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Name != "", "name", "must be provided")
	v.Check(input.TeacherID > 0, "teacher_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	class := &data.Class{
		Name:     input.Name,
		Teacher:  &data.User{ID: input.TeacherID},
		Archived: false,
	}

	teacher, err := app.models.Users.GetUserByID(class.Teacher.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if teacher.Role != data.RoleAdministrator && teacher.Role != data.RoleTeacher {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	err = app.models.Classes.InsertClass(class)
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
		Name      *string `json:"name"`
		TeacherID *int    `json:"teacher_id"`
		Archived  *bool   `json:"archived"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name != nil {
		class.Name = *input.Name
	}
	if input.TeacherID != nil {
		class.Teacher.ID = *input.TeacherID
	}
	if input.Archived != nil {
		class.Archived = *input.Archived
	}

	v := validator.NewValidator()
	v.Check(class.Name != "", "name", "must be provided")
	v.Check(class.Teacher.ID > 0, "teacher_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	teacher, err := app.models.Users.GetUserByID(class.Teacher.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if teacher.Role != data.RoleAdministrator && teacher.Role != data.RoleTeacher {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	err = app.models.Classes.UpdateClass(class)
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

func (app *application) getClassForStudent(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	class, err := app.models.Classes.GetClassForUserID(user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoClassForUser):
			class = nil
		default:
			app.writeInternalServerError(w, r, err)
			return
		}

	}

	err = app.outputJSON(w, http.StatusOK, envelope{"class": class})
	if err != nil {
		app.writeInternalServerError(w, r, err)
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

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		if class.Teacher.ID != sessionUser.ID {
			app.notAllowed(w, r)
			return
		}
	case data.RoleParent:
		ok, err := app.models.Classes.DoesParentHaveChildInClass(sessionUser.ID, class.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		ok, err := app.models.Classes.IsUserInClass(sessionUser.ID, class.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
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

func (app *application) setClassForStudent(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	var input struct {
		ClassID *int `json:"class_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.ClassID != nil, "class_id", "must be provided")
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	if *input.ClassID < 1 {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchClass.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(*input.ClassID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if class.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrClassArchived.Error())
		return
	}

	err = app.models.Classes.SetClassIDForUserID(user.ID, class.ID)
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
