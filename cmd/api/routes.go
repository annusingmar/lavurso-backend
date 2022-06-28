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

	// get student's class
	mux.HandlerFunc(http.MethodGet, "/students/:id/class", app.getClassForStudent)

	// set student's class
	mux.HandlerFunc(http.MethodPut, "/students/:id/class", app.setClassForStudent)

	// get students in class
	mux.HandlerFunc(http.MethodGet, "/classes/:id/students", app.getStudentsInClass)

	// list all subjects
	mux.HandlerFunc(http.MethodGet, "/subjects", app.listAllSubjects)

	// get subject by id
	mux.HandlerFunc(http.MethodGet, "/subjects/:id", app.getSubject)

	// create subject
	mux.HandlerFunc(http.MethodPost, "/subjects", app.createSubject)

	// update subject
	mux.HandlerFunc(http.MethodPatch, "/subjects/:id", app.updateSubject)

	// create journal
	mux.HandlerFunc(http.MethodPost, "/journals", app.createJournal)

	// update journal
	mux.HandlerFunc(http.MethodPatch, "/journals/:id", app.updateJournal)

	// delete journal
	mux.HandlerFunc(http.MethodDelete, "/journals/:id", app.deleteJournal)

	// get all journals
	mux.HandlerFunc(http.MethodGet, "/journals", app.listAllJournals)

	// get journal by id
	mux.HandlerFunc(http.MethodGet, "/journals/:id", app.getJournal)

	// get journals for teacher
	mux.HandlerFunc(http.MethodGet, "/teachers/:id/journals", app.getJournalsForTeacher)

	// get users for journal
	mux.HandlerFunc(http.MethodGet, "/journals/:id/students", app.getStudentsForJournal)

	// get journals for user
	mux.HandlerFunc(http.MethodGet, "/students/:id/journals", app.getJournalsForStudent)

	// add user to journal
	mux.HandlerFunc(http.MethodPost, "/students/:id/journals", app.addStudentToJournal)

	// remove user from journal
	mux.HandlerFunc(http.MethodDelete, "/students/:id/journals", app.removeStudentFromJournal)

	// create lesson
	mux.HandlerFunc(http.MethodPost, "/lessons", app.createLesson)

	// get lesson by id
	mux.HandlerFunc(http.MethodGet, "/lessons/:id", app.getLesson)

	// update lesson
	mux.HandlerFunc(http.MethodPatch, "/lessons/:id", app.updateLesson)

	// get lessons for journal
	mux.HandlerFunc(http.MethodGet, "/journals/:id/lessons", app.getLessonsForJournal)

	// get assignment by id
	mux.HandlerFunc(http.MethodGet, "/assignments/:id", app.getAssignment)

	// create assignment
	mux.HandlerFunc(http.MethodPost, "/assignments", app.createAssignment)

	// update assignment
	mux.HandlerFunc(http.MethodPatch, "/assignments/:id", app.updateAssignment)

	// delete assignment
	mux.HandlerFunc(http.MethodDelete, "/assignments/:id", app.deleteAssignment)

	return mux
}
