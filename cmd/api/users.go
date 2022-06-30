package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) listAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.models.Users.AllUsers()
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"users": users})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	user := &data.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: data.Password{Plaintext: input.Password},
		Role:     input.Role,
	}

	app.models.Users.ValidateUser(v, user)
	app.models.Users.ValidatePassword(v, user)
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	user.Password.Hashed, err = app.models.Users.HashPassword(user.Password.Plaintext)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	user.CreatedAt = time.Now().UTC()
	user.Active = true
	user.Version = 1

	err = app.models.Users.InsertUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEmailAlreadyExists):
			app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"user": user})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
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

	err = app.outputJSON(w, http.StatusOK, envelope{"user": user})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
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

	var input struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
		Active   *bool   `json:"active"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		user.Password.Plaintext = *input.Password
		app.models.Users.ValidatePassword(v, user)
		user.Password.Hashed, err = app.models.Users.HashPassword(user.Password.Plaintext)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}
	if input.Active != nil {
		user.Active = *input.Active
	}

	app.models.Users.ValidateUser(v, user)
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.models.Users.UpdateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
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
