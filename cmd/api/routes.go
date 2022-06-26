package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	mux := httprouter.New()
	mux.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	// administrator
	mux.HandlerFunc(http.MethodGet, "/roles", app.listRoles)
	mux.HandlerFunc(http.MethodGet, "/users", app.listAllUsers)
	mux.HandlerFunc(http.MethodPost, "/users/create", app.createUser)
	mux.HandlerFunc(http.MethodPatch, "/users/:id/update", app.updateUser)
	mux.HandlerFunc(http.MethodGet, "/users/:id/class", app.getClassForUser)
	mux.HandlerFunc(http.MethodGet, "/classes/:id/users", app.getUsersInClass)
	return mux
}
