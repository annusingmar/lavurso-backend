package main

import (
	"net/http"
	"time"

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
	// phone, address ja birthdate on pointerid, kuna
	// siis saab vorrelda nil-iga, st saab defaultida tyhjale
	var input struct {
		Name      string     `json:"name"`
		Email     string     `json:"email"`
		Password  string     `json:"password"`
		Phone     *string    `json:"phone,omitempty"`
		Address   *string    `json:"address,omitempty"`
		BirthDate *time.Time `json:"birth_date,omitempty"`
		Role      int        `json:"role"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeBadRequestError(w, r, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.Name != "", "name", "must be provided")
	v.Check(input.Email != "", "email", "must be provided")
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(input.Phone != nil, "phone", "must be provided")
	v.Check(input.Address != nil, "address", "must be provided")
	v.Check(input.BirthDate != nil, "birth_date", "must be provided")
	v.Check(input.Role != 0, "role", "must be provided")

	if !v.Valid() {
		app.writeBadRequestError(w, r, v.Errors)
		return
	}

	// TODO: sisestada kasutaja

}
