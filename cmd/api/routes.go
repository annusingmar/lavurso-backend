package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.StripSlashes)

	mux.MethodNotAllowed(app.methodNotAllowed)
	mux.NotFound(app.notFound)

	// list all users
	mux.Get("/users", app.listAllUsers)

	// create new user
	mux.Post("/users", app.createUser)

	// get user by id
	mux.Get("/users/{id}", app.getUser)

	// update user
	mux.Patch("/users/{id}", app.updateUser)

	// list all classes
	mux.Get("/classes", app.listAllClasses)

	// create new class
	mux.Post("/classes", app.createClass)

	// get class by id
	mux.Get("/classes/{id}", app.getClass)

	// update class
	mux.Patch("/classes/{id}", app.updateClass)

	// get student's class
	mux.Get("/students/{id}/class", app.getClassForStudent)

	// set student's class
	mux.Put("/students/{id}/class", app.setClassForStudent)

	// get students in class
	mux.Get("/classes/{id}/students", app.getStudentsInClass)

	// list all subjects
	mux.Get("/subjects", app.listAllSubjects)

	// get subject by id
	mux.Get("/subjects/{id}", app.getSubject)

	// create subject
	mux.Post("/subjects", app.createSubject)

	// update subject
	mux.Patch("/subjects/{id}", app.updateSubject)

	// create journal
	mux.Post("/journals", app.createJournal)

	// update journal
	mux.Patch("/journals/{id}", app.updateJournal)

	// delete journal
	mux.Delete("/journals/{id}", app.deleteJournal)

	// get all journals
	mux.Get("/journals", app.listAllJournals)

	// get journal by id
	mux.Get("/journals/{id}", app.getJournal)

	// get journals for teacher
	mux.Get("/teachers/{id}/journals", app.getJournalsForTeacher)

	// get users for journal
	mux.Get("/journals/{id}/students", app.getStudentsForJournal)

	// get journals for user
	mux.Get("/students/{id}/journals", app.getJournalsForStudent)

	// add user to journal
	mux.Post("/students/{id}/journals", app.addStudentToJournal)

	// remove user from journal
	mux.Delete("/students/{id}/journals", app.removeStudentFromJournal)

	// create lesson
	mux.Post("/lessons", app.createLesson)

	// get lesson by id
	mux.Get("/lessons/{id}", app.getLesson)

	// update lesson
	mux.Patch("/lessons/{id}", app.updateLesson)

	// get lessons for journal
	mux.Get("/journals/{id}/lessons", app.getLessonsForJournal)

	// get assignment by id
	mux.Get("/assignments/{id}", app.getAssignment)

	// create assignment
	mux.Post("/assignments", app.createAssignment)

	// update assignment
	mux.Patch("/assignments/{id}", app.updateAssignment)

	// delete assignment
	mux.Delete("/assignments/{id}", app.deleteAssignment)

	// get all assignments for journal
	mux.Get("/journals/{id}/assignments", app.getAssignmentsForJournal)

	// get all assignments for student
	mux.Get("/students/{id}/assignments", app.getAssignmentsForStudent)

	// set assignment done for student
	mux.Put("/students/{sid}/assignments/{aid}/done", app.setAssignmentDoneForStudent)

	// remove assignment done for student
	mux.Delete("/students/{sid}/assignments/{aid}/done", app.removeAssignmentDoneForStudent)

	// get all grades
	mux.Get("/grades", app.listAllGrades)

	// get grade by id
	mux.Get("/grades/{id}", app.getGrade)

	// create grade
	mux.Post("/grades", app.createGrade)

	// update grade
	mux.Patch("/grades/{id}", app.updateGrade)

	// get mark by id
	mux.Get("/marks/{id}", app.getMark)

	// get current marks for student
	mux.Get("/students/{id}/marks", app.getMarksForStudent)

	// get current marks for journal
	mux.Get("/journals/{id}/marks", app.getMarksForJournal)

	// get current marks for student's journal
	mux.Get("/students/{sid}/journals/{jid}/marks", app.getMarksForStudentsJournal)

	// add mark
	mux.Post("/marks", app.addMark)

	// delete mark
	mux.Delete("/marks/{id}", app.deleteMark)

	// update mark
	mux.Patch("/marks/{id}", app.updateMark)

	// get previous marks for mark
	mux.Get("/marks/{id}/previous", app.getPreviousMarksForMark)

	// get absences for student
	mux.Get("/students/{id}/absences", app.getAbsencesForStudent)

	// excuse absence for student
	mux.Post("/students/{id}/excuses", app.excuseAbsenceForStudent)

	// delete excuse for student
	mux.Delete("/students/{sid}/excuses/{eid}", app.deleteAbsenceExcuseForStudent)

	return mux
}
