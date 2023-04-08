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
	mux.Use(app.log)

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
			mux.Patch("/users/{id}", app.updateUserAdmin)

			// get class by id
			mux.Get("/classes/{id}", app.getClass)

			// create new class
			mux.Post("/classes", app.createClass)

			// update class
			mux.Patch("/classes/{id}", app.updateClass)

			// create subject
			mux.Post("/subjects", app.createSubject)

			// update subject
			mux.Patch("/subjects/{id}", app.updateSubject)

			// delete subject
			mux.Delete("/subjects/{id}", app.deleteSubject)

			// get grade by id
			mux.Get("/grades/{id}", app.getGrade)

			// create grade
			mux.Post("/grades", app.createGrade)

			// get all groups
			mux.Get("/groups", app.getAllGroups)

			// create group
			mux.Post("/groups", app.createGroup)

			// update grade
			mux.Patch("/grades/{id}", app.updateGrade)

			// get group by id
			mux.Get("/groups/{id}", app.getGroup)

			// update group
			mux.Patch("/groups/{id}", app.updateGroup)

			// delete group
			mux.Delete("/groups/{id}", app.deleteGroup)

			// add users to group
			mux.Post("/groups/{id}/users", app.addUsersToGroup)

			// delete users from groups
			mux.Delete("/groups/{id}/users", app.removeUsersFromGroup)

			// get users by group id
			mux.Get("/groups/{id}/users", app.getUsersForGroup)

			// get all journals
			mux.Get("/journals", app.listAllJournals)

			// delete journal
			mux.Delete("/journals/{id}", app.deleteJournal)

			// add parent to student
			mux.Put("/students/{id}/parents", app.addParentToStudent)

			// remove parent from student
			mux.Delete("/students/{id}/parents", app.removeParentFromStudent)

			// new year
			mux.Post("/years/new", app.newYear)

			mux.Get("/classes/{id}/years", app.getYearsForClass)

			mux.Put("/classes/{id}/years", app.setYearsForClass)

			// get all sessions for user
			mux.Get("/users/{id}/sessions", app.allSessionsForUser)

			// delete all sesions for user
			mux.Delete("/users/{id}/sessions", app.expireAllSessionsForUser)

			mux.Get("/logs", app.getLogs)
		})

		// requires at least role 'teacher'
		mux.Group(func(mux chi.Router) {
			mux.Use(app.requireTeacher)

			// create journal
			mux.Post("/journals", app.createJournal)

			// get journal by id
			mux.Get("/journals/{id}", app.getJournal)

			// update journal
			mux.Patch("/journals/{id}", app.updateJournal)

			// get journals for teacher
			mux.Get("/teachers/{id}/journals", app.getJournalsForTeacher)

			// get classes for teacher
			mux.Get("/teachers/{id}/classes", app.getClassesForTeacher)

			// get students in class
			mux.Get("/classes/{id}/students", app.getStudentsInClass)

			// get users for journal
			mux.Get("/journals/{id}/students", app.getStudentsForJournal)

			// add users to journal
			mux.Post("/journals/{id}/students", app.addStudentsToJournal)

			// remove user from journal
			mux.Delete("/journals/{id}/students", app.removeStudentFromJournal)

			// get lesson by id
			mux.Get("/lessons/{id}", app.getLesson)

			// create lesson
			mux.Post("/lessons", app.createLesson)

			// update lesson
			mux.Patch("/lessons/{id}", app.updateLesson)

			// delete lesson
			mux.Delete("/lessons/{id}", app.deleteLesson)

			// get all grades
			mux.Get("/grades", app.listAllGrades)

			// get lessons for journal
			mux.Get("/journals/{id}/lessons", app.getLessonsForJournal)

			// get assignment by id
			mux.Get("/assignments/{id}", app.getAssignment)

			// get all assignments for journal
			mux.Get("/journals/{id}/assignments", app.getAssignmentsForJournal)

			// create assignment
			mux.Post("/assignments", app.createAssignment)

			// update assignment
			mux.Patch("/assignments/{id}", app.updateAssignment)

			// delete assignment
			mux.Delete("/assignments/{id}", app.deleteAssignment)

			// get students and marks for lesson
			mux.Get("/lessons/{id}/marks", app.getMarksForLesson)

			// save marks for lesson
			mux.Patch("/lessons/{id}/marks", app.setMarksForLesson)

			// get course + all lessons marks for course
			mux.Get("/journals/{jid}/courses/{course}/marks", app.getMarksForCourse)

			// save marks for course
			mux.Patch("/journals/{jid}/courses/{course}/marks", app.setMarksForCourse)

			// get subject + all course marks for journal
			mux.Get("/journals/{jid}/subject/marks", app.getMarksForJournalSubject)

			// save marks for journal's subject
			mux.Patch("/journals/{jid}/subject/marks", app.setMarksForJournalSubject)

			// list all subjects
			mux.Get("/subjects", app.listAllSubjects)

			// list all classes
			mux.Get("/classes", app.listAllClasses)
		})

		// search for user with query param 'name' (minimum 4 characters)
		mux.Get("/users/search", app.searchUser)

		// get all assignments for student
		mux.Get("/students/{id}/assignments", app.getAssignmentsForStudent)

		// set assignment done for student
		mux.Put("/students/{sid}/assignments/{aid}/done", app.setAssignmentDoneForStudent)

		// remove assignment done for student
		mux.Delete("/students/{sid}/assignments/{aid}/done", app.removeAssignmentDoneForStudent)

		// get current marks for student
		mux.Get("/students/{id}/marks", app.getMarksForStudent)

		// get all grades for student
		mux.Get("/students/{id}/grades", app.getGradesByYearForStudent)

		// get lessons and marks for student's journal
		mux.Get("/students/{sid}/journals/{jid}/lessons", app.getLessonsForStudentsJournalsCourse)

		// excuse absence for student
		mux.Post("/absences/{id}/excuse", app.excuseAbsenceForStudent)

		// delete excuse for student
		mux.Delete("/absences/{id}/excuse", app.deleteExcuseForStudent)

		// get groups by user id
		mux.Get("/users/{id}/groups", app.getGroupsForUser)

		// get all threads for user
		mux.Get("/me/threads", app.getThreadsForUser)

		// does user have unread
		mux.Get("/me/unread", app.userHasUnread)

		// create thread
		mux.Post("/threads", app.createThread)

		// lock thread
		mux.Put("/threads/{id}/lock", app.lockThread)

		// unlock thread
		mux.Put("/threads/{id}/unlock", app.unlockThread)

		// delete thread
		mux.Delete("/threads/{id}", app.deleteThread)

		// add members to thread
		mux.Put("/threads/{id}/members", app.addMembersToThread)

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

		// delete session by id
		mux.Delete("/sessions/{id}", app.expireSession)

		// get student by id
		mux.Get("/students/{id}", app.getStudent)

		// get latest marks/lessons for student
		mux.Get("/students/{id}/latest", app.getLatestMarksLessonsForStudent)

		// get user by id
		mux.Get("/users/{id}", app.getUser)

		// update own user
		mux.Put("/me", app.updateUser)

		// change own password
		mux.Post("/me/password", app.changeUserPassword)

		// get user, children, year
		mux.Get("/me", app.myInfo)

		// all years
		mux.Get("/years", app.getAllYears)

		// years for student's class
		mux.Get("/students/{id}/years", app.getYearsForStudent)

		// start 2fa (generate secret)
		mux.Post("/me/2fa", app.start2FA)

		// enable 2fa
		mux.Post("/me/2fa/finish", app.enable2FA)

		// disable 2fa
		mux.Delete("/me/2fa", app.disable2FA)

		// logout
		mux.Post("/me/logout", app.logout)
	})

	return mux
}
