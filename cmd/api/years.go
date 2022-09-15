package main

import "net/http"

func (app *application) getAllYears(w http.ResponseWriter, r *http.Request) {
	years, err := app.models.Years.ListAllYears()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
