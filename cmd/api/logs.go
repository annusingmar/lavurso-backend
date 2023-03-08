package main

import (
	"net/http"
	"strconv"
)

func (app *application) getLogs(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 0 {
		limit = 50
	}

	logs, err := app.models.Logs.AllLogs(page, limit, search)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"result": logs})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
