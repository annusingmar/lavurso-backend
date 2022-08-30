package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) listAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.models.Users.AllUsers()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"users": users})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) searchUser(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))

	if utf8.RuneCountInString(name) < 4 {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "not enough characters provided")
		return
	}

	result, err := app.models.Users.SearchUser(name)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"result": result})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string     `json:"name"`
		Email       string     `json:"email"`
		Password    string     `json:"password"`
		PhoneNumber *string    `json:"phone_number"`
		IdCode      *int64     `json:"id_code"`
		BirthDate   *data.Date `json:"birth_date"`
		Role        string     `json:"role"`
		ClassID     *int       `json:"class_id"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Name != "", "name", "must be provided")
	v.Check(input.Email != "", "email", "must be provided")
	v.Check(data.EmailRegex.MatchString(input.Email), "email", "must be a valid email address")
	v.Check(input.PhoneNumber == nil || *input.PhoneNumber != "", "phone_number", "must not be empty")
	v.Check(input.IdCode == nil || len(fmt.Sprint(*input.IdCode)) == 11, "id_code", "must be 11 digits long")

	if input.BirthDate == nil || input.BirthDate.Time == nil {
		input.BirthDate = new(data.Date)
	} else {
		if input.BirthDate.Time.After(time.Now().UTC()) {
			app.writeErrorResponse(w, r, http.StatusBadRequest, "time is in the future")
			return
		}
	}

	v.Check(input.Role == data.RoleAdministrator || input.Role == data.RoleTeacher || input.Role == data.RoleParent || input.Role == data.RoleStudent, "role", "must be valid role")

	v.Check(input.Password != "", "password", "must be provided")

	if input.Role == data.RoleStudent {
		v.Check(input.ClassID != nil, "class_id", "must be provided")
	}

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	var classField *data.Class
	if input.Role == data.RoleStudent {
		class, err := app.models.Classes.GetClassByID(*input.ClassID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoSuchClass):
				app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}
		classField = &data.Class{ID: class.ID}
	} else {
		classField = new(data.Class)
	}

	user := &data.User{
		Name:        &input.Name,
		Email:       &input.Email,
		Password:    data.Password{Plaintext: input.Password},
		PhoneNumber: input.PhoneNumber,
		IdCode:      input.IdCode,
		BirthDate:   input.BirthDate,
		Role:        &input.Role,
		Class:       classField,
	}

	user.Password.Hashed, err = app.models.Users.HashPassword(user.Password.Plaintext)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	user.CreatedAt = helpers.ToPtr(time.Now().UTC())

	err = app.models.Users.InsertUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEmailAlreadyExists) || errors.Is(err, data.ErrIDCodeAlreadyExists):
			app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	if userID != *sessionUser.ID && *sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var user *data.User

	user, err = app.models.Users.GetUserByID(userID)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"user": user})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) updateUserAdmin(w http.ResponseWriter, r *http.Request) {
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

	var input struct {
		Name        *string    `json:"name"`
		Email       *string    `json:"email"`
		Password    *string    `json:"password"`
		PhoneNumber *string    `json:"phone_number"`
		IdCode      *int64     `json:"id_code"`
		BirthDate   *data.Date `json:"birth_date"`
		ClassID     *int       `json:"class_id"`
		Active      *bool      `json:"active"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Name == nil || *input.Name != "", "name", "must not be empty")
	v.Check(input.Email == nil || *input.Email != "", "email", "must not be empty")
	v.Check(input.Email == nil || data.EmailRegex.MatchString(*input.Email), "email", "must be a valid email address")
	v.Check(input.Password == nil || *input.Password != "", "password", "must not be empty")
	v.Check(input.PhoneNumber == nil || *input.PhoneNumber != "", "phone_number", "must not be empty")
	v.Check(input.IdCode == nil || len(fmt.Sprint(*input.IdCode)) == 11, "id_code", "must be 11 digits long")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	if input.Name != nil {
		user.Name = input.Name
	}

	if input.Email != nil {
		user.Email = input.Email
	}

	if input.Password != nil {
		user.Password.Hashed, err = app.models.Users.HashPassword(*input.Password)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	user.PhoneNumber = input.PhoneNumber
	user.IdCode = input.IdCode

	if input.BirthDate != nil && input.BirthDate.Time != nil {
		if input.BirthDate.Time.After(time.Now().UTC()) {
			app.writeErrorResponse(w, r, http.StatusBadRequest, "time is in the future")
			return
		}
		user.BirthDate = input.BirthDate
	} else {
		user.BirthDate = new(data.Date)
	}

	if input.ClassID != nil && *user.Role == data.RoleStudent {
		class, err := app.models.Classes.GetClassByID(*input.ClassID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoSuchClass):
				app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}
		user.Class = &data.Class{ID: class.ID}
	}

	err = app.models.Users.UpdateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEmailAlreadyExists) || errors.Is(err, data.ErrIDCodeAlreadyExists):
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

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	if *user.ID != *sessionUser.ID {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		Email       string  `json:"email"`
		PhoneNumber *string `json:"phone_number"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Email != "", "email", "must not be empty")
	v.Check(input.PhoneNumber == nil || *input.PhoneNumber != "", "phone_number", "must not be empty")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	user.Email = &input.Email
	user.PhoneNumber = input.PhoneNumber

	err = app.models.Users.UpdateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEmailAlreadyExists):
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

func (app *application) changeUserPassword(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	if *user.ID != *sessionUser.ID {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.CurrentPassword != "", "current_password", "must not be empty")
	v.Check(input.NewPassword != "", "new_password", "must not be empty")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	correct, err := data.ComparePassword(user.Password.Hashed, input.CurrentPassword)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	if !correct {
		app.writeErrorResponse(w, r, http.StatusConflict, ErrInvalidCredentials.Error())
		return
	}

	user.Password.Plaintext = input.NewPassword
	user.Password.Hashed, err = app.models.Users.HashPassword(user.Password.Plaintext)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Users.UpdateUser(user)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Sessions.RemoveAllSessionsByUserIDExceptOne(*user.ID, *sessionUser.SessionID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getStudent(w http.ResponseWriter, r *http.Request) {
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

	switch *sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		if *student.Class.Teacher.ID != *sessionUser.ID {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	parents, err := app.models.Users.GetParentsForChild(*student.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"student": student, "parents": parents})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addParentToStudent(w http.ResponseWriter, r *http.Request) {
	studentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if studentID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	student, err := app.models.Users.GetUserByID(studentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *student.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	var input struct {
		ParentID int `json:"parent_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.ParentID > 0, "parent_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	parent, err := app.models.Users.GetUserByID(input.ParentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *parent.Role != data.RoleParent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAParent.Error())
		return
	}

	err = app.models.Users.AddParentToChild(*parent.ID, *student.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeParentFromStudent(w http.ResponseWriter, r *http.Request) {
	studentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if studentID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	student, err := app.models.Users.GetUserByID(studentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *student.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	var input struct {
		ParentID int `json:"parent_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.ParentID > 0, "parent_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	parent, err := app.models.Users.GetUserByID(input.ParentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *parent.Role != data.RoleParent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAParent.Error())
		return
	}

	ok, err := app.models.Users.IsUserParentOfStudent(*parent.ID, *student.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	if !ok {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNoSuchParentForUser.Error())
		return
	}

	err = app.models.Users.RemoveParentFromChild(*parent.ID, *student.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
