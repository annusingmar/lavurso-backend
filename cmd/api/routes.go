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
	mux.Use(app.authenticateSession)

	mux.MethodNotAllowed(app.methodNotAllowed)
	mux.NotFound(app.notFound)

	// authenticate user
	mux.Post("/authenticate", app.authenticateUser)

	// requires auth
	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)

		// requires role 'admin'
		mux.Group(func(mux chi.Router) {
			mux.Use(app.requireAdministrator)

			// list all users
			mux.Get("/users", app.listAllUsers)

			// create new user
			mux.Post("/users", app.createUser)

			// update user
			mux.Patch("/users/{id}", app.updateUser)

			// create new class
			mux.Post("/classes", app.createClass)

			// update class
			mux.Patch("/classes/{id}", app.updateClass)

			// set student's class
			mux.Put("/students/{id}/class", app.setClassForStudent)

			// create subject
			mux.Post("/subjects", app.createSubject)

			// update subject
			mux.Patch("/subjects/{id}", app.updateSubject)

			// create grade
			mux.Post("/grades", app.createGrade)

			// get all groups
			mux.Get("/groups", app.getAllGroups)

			// create group
			mux.Post("/groups", app.createGroup)

			// update grade
			mux.Patch("/grades/{id}", app.updateGrade)

			// update group
			mux.Patch("/groups/{id}", app.updateGroup)

			// delete group
			mux.Delete("/groups/{id}", app.removeGroup)

			// add users to group
			mux.Post("/groups/{id}/users", app.addUsersToGroup)

			// delete users from groups
			mux.Delete("/groups/{id}/users", app.removeUsersFromGroup)

			// get all journals
			mux.Get("/journals", app.listAllJournals)

			// add parent to student
			mux.Put("/students/{id}/parents", app.addParentToStudent)

			// remove parent from student
			mux.Delete("/students/{id}/parents", app.removeParentFromStudent)

		})

		// requires at least role 'teacher'
		mux.Group(func(mux chi.Router) {
			mux.Use(app.requireTeacher)

			// list all classes
			mux.Get("/classes", app.listAllClasses)

			// create journal
			mux.Post("/journals", app.createJournal)

			// update journal
			mux.Patch("/journals/{id}", app.updateJournal)

			// delete journal
			mux.Delete("/journals/{id}", app.deleteJournal)

			// get journals for teacher
			mux.Get("/teachers/{id}/journals", app.getJournalsForTeacher)

			// get users for journal
			mux.Get("/journals/{id}/students", app.getStudentsForJournal)

			// add user to journal
			mux.Post("/journals/{id}/students", app.addStudentToJournal)

			// remove user from journal
			mux.Delete("/journals/{id}/students", app.removeStudentFromJournal)

			// create lesson
			mux.Post("/lessons", app.createLesson)

			// update lesson
			mux.Patch("/lessons/{id}", app.updateLesson)

			// create assignment
			mux.Post("/assignments", app.createAssignment)

			// update assignment
			mux.Patch("/assignments/{id}", app.updateAssignment)

			// delete assignment
			mux.Delete("/assignments/{id}", app.deleteAssignment)

			// get current marks for journal
			mux.Get("/journals/{id}/marks", app.getMarksForJournal)

			// add mark
			mux.Post("/marks", app.addMark)

			// delete mark
			mux.Delete("/marks/{id}", app.deleteMark)

			// update mark
			mux.Patch("/marks/{id}", app.updateMark)
		})

		mux.Get("/users/search", app.searchUser)

		// get user by id
		mux.Get("/users/{id}", app.getUser)

		// get class by id
		mux.Get("/classes/{id}", app.getClass)

		// get student's class
		mux.Get("/students/{id}/class", app.getClassForStudent)

		// get students in class
		mux.Get("/classes/{id}/students", app.getStudentsInClass)

		// list all subjects
		mux.Get("/subjects", app.listAllSubjects)

		// get subject by id
		mux.Get("/subjects/{id}", app.getSubject)

		// get journal by id
		mux.Get("/journals/{id}", app.getJournal)

		// get journals for user
		mux.Get("/students/{id}/journals", app.getJournalsForStudent)

		// get lesson by id
		mux.Get("/lessons/{id}", app.getLesson)

		// get lessons for journal
		mux.Get("/journals/{id}/lessons", app.getLessonsForJournal)

		// get assignment by id
		mux.Get("/assignments/{id}", app.getAssignment)

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

		// get mark by id
		mux.Get("/marks/{id}", app.getMark)

		// get current marks for student
		mux.Get("/students/{id}/marks", app.getMarksForStudent)

		// get current marks for student's journal
		mux.Get("/students/{sid}/journals/{jid}/marks", app.getMarksForStudentsJournal)

		// get previous marks for mark
		mux.Get("/marks/{id}/previous", app.getPreviousMarksForMark)

		// get absences for student
		mux.Get("/students/{id}/absences", app.getAbsencesForStudent)

		// excuse absence for student
		mux.Post("/students/{id}/excuses", app.excuseAbsenceForStudent)

		// delete excuse for student
		mux.Delete("/students/{sid}/excuses/{eid}", app.deleteAbsenceExcuseForStudent)

		// get group by id
		mux.Get("/groups/{id}", app.getGroup)

		// get groups by user id
		mux.Get("/users/{id}/groups", app.getGroupsForUser)

		// get users by group id
		mux.Get("/groups/{id}/users", app.getUsersForGroup)

		// get all threads for user
		mux.Get("/users/{id}/threads", app.getThreadsForUser)

		// create thread
		mux.Post("/threads", app.createThread)

		// lock thread
		mux.Put("/threads/{id}/lock", app.lockThread)

		// unlock thread
		mux.Put("/threads/{id}/unlock", app.unlockThread)

		// delete thread
		mux.Delete("/threads/{id}", app.deleteThread)

		// add members to thread
		mux.Put("/threads/{id}/members", app.addNewMembersToThread)

		// remove members from thread
		mux.Delete("/threads/{id}/members", app.removeMembersFromThread)

		// get members in threads
		mux.Get("/threads/{id}/members", app.getThreadMembers)

		// create message
		mux.Post("/threads/{id}/messages", app.createMessage)

		// edit message
		mux.Put("/messages/{id}", app.updateMessage)

		// delete message
		mux.Delete("/messages/{id}", app.deleteMessage)

		// get thread by id
		mux.Get("/threads/{id}", app.getThread)

		// get all sessions for user
		mux.Get("/users/{id}/sessions", app.allSessionsForUser)

		// delete all sesions for user
		mux.Delete("/users/{id}/sessions", app.removeAllSessionsForUser)

		// delete session by id
		mux.Delete("/sessions/{id}", app.removeSession)

		// get parents for student
		mux.Get("/students/{id}/parents", app.getParentsForStudent)
	})

	return mux
}
