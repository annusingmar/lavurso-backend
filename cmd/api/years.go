package main

import (
	"net/http"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

func (app *application) getAllYears(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var err error
	var years []*data.Year

	if *sessionUser.Role == data.RoleAdministrator && r.URL.Query().Get("stats") == "true" {
		years, err = app.models.Years.ListAllYearsWithStats()
	} else {
		years, err = app.models.Years.ListAllYears()

	}

	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
