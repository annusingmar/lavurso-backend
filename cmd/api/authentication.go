package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/annusingmar/lavurso-backend/internal/validator"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

func (app *application) authenticateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Email != "", "email", "must be provided")
	v.Check(input.Password != "", "password", "must be provided")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	user, err := app.models.Users.GetUserByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusForbidden, ErrInvalidCredentials.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	correct, err := user.Password.Validate(input.Password)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	if !correct {
		app.writeErrorResponse(w, r, http.StatusForbidden, ErrInvalidCredentials.Error())
		return
	}

	ip, err := app.getIP(r)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	currentTime := time.Now().UTC()

	session := &model.Sessions{
		UserID:       &user.ID,
		Token:        new(types.Token),
		Expires:      helpers.ToPtr(currentTime.Add(72 * time.Hour)),
		LoginIP:      &ip,
		LoginBrowser: helpers.ToPtr(r.UserAgent()),
		LoggedIn:     &currentTime,
		LastSeen:     &currentTime,
	}

	err = session.Token.NewToken()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Sessions.InsertSession(session)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusAccepted, envelope{"session": session})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}
