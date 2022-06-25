package main

import "net/http"

func (app *application) listAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.models.Users.AllUsers()
	if err != nil {
		panic(err)
	}
	err = app.outputJSON(w, http.StatusOK, envelope{"users": users})
	if err != nil {
		panic(err)
	}
}
