package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
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
		Role     int    `json:"role"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeBadRequestError(w, r, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.Name != "", "name", "must be provided")
	v.Check(input.Email != "", "email", "must be provided")
	v.Check(data.EmailRegex.MatchString(input.Email), "email", "must be a valid email address")
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(input.Role > 0 && input.Role < 4, "role", "must be {1,2,3}")

	if !v.Valid() {
		app.writeBadRequestError(w, r, v.Errors)
		return
	}

	user := &data.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: data.Password{Plaintext: input.Password},
		Role:     input.Role,
	}

	user.Password.Hashed, err = app.models.Users.HashPassword(user.Password.Plaintext)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	user.CreatedAt = time.Now().UTC()
	user.Version = 1

	err = app.models.Users.InsertUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEmailAlreadyExists):
			app.writeErrorResponse(w, r, http.StatusConflict, data.ErrEmailAlreadyExists.Error())
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
