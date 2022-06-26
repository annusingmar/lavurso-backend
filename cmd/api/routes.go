package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	mux := httprouter.New()
	mux.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	// list all roles
	mux.HandlerFunc(http.MethodGet, "/roles", app.listRoles)

	// list all users
	mux.HandlerFunc(http.MethodGet, "/users", app.listAllUsers)

	// create new user
	mux.HandlerFunc(http.MethodPost, "/users", app.createUser)

	// get user by id
	mux.HandlerFunc(http.MethodGet, "/users/:id", app.getUser)

	// update user
	mux.HandlerFunc(http.MethodPatch, "/users/:id", app.updateUser)

	// list all classes
	mux.HandlerFunc(http.MethodGet, "/classes", app.listAllClasses)

	// create new class
	mux.HandlerFunc(http.MethodPost, "/classes", app.createClass)

	// get class by id
	mux.HandlerFunc(http.MethodGet, "/classes/:id", app.getClass)

	// update class
	mux.HandlerFunc(http.MethodPatch, "/classes/:id", app.updateClass)

	// get user's class
	mux.HandlerFunc(http.MethodGet, "/users/:id/class", app.getClassForUser)

	// set user's class
	mux.HandlerFunc(http.MethodPut, "/users/:id/class", app.setClassForUser)

	// get users in class
	mux.HandlerFunc(http.MethodGet, "/classes/:id/users", app.getUsersInClass)

	// list all subjects
	mux.HandlerFunc(http.MethodGet, "/subjects", app.listAllSubjects)

	// get subject by id
	mux.HandlerFunc(http.MethodGet, "/subjects/:id", app.getSubject)

	// create subject
	mux.HandlerFunc(http.MethodPost, "/subjects", app.createSubject)

	// update subject
	mux.HandlerFunc(http.MethodPatch, "/subjects/:id", app.updateSubject)

	return mux
}
