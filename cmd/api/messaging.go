package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) createThread(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Body    string `json:"body"`
		UserIDs []int  `json:"user_ids"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	thread := &data.Thread{
		UserID:    1, // to change
		Title:     input.Title,
		Body:      input.Body,
		Locked:    false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	v.Check(thread.Title != "", "title", "must be present")
	v.Check(thread.Body != "", "body", "must be present")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	allUserIDs, err := app.models.Users.GetAllUserIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	badIDs := helpers.VerifyExistsInSlice(input.UserIDs, allUserIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchUsers.Error(), badIDs))
		return
	}

	err = app.models.Messaging.InsertThread(thread)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	for _, id := range input.UserIDs {
		err = app.models.Messaging.AddUserToThread(id, thread.ID)
		if err != nil && !errors.Is(err, data.ErrUserAlreadyInThread) {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"thread": thread})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) updateThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if threadID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchThread.Error())
		return
	}

	thread, err := app.models.Messaging.GetThreadByID(threadID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchThread):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Title != "", "title", "must be present")
	v.Check(input.Body != "", "body", "must be present")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	if input.Title != thread.Title || input.Body != thread.Body {
		thread.Title = input.Title
		thread.Body = input.Body
		thread.UpdatedAt = time.Now().UTC()

		err = app.models.Messaging.UpdateThread(thread)
		if err != nil {
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"thread": thread})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) deleteThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if threadID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchThread.Error())
		return
	}

	thread, err := app.models.Messaging.GetThreadByID(threadID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchThread):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Messaging.DeleteThread(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) lockThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if threadID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchThread.Error())
		return
	}

	thread, err := app.models.Messaging.GetThreadByID(threadID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchThread):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if thread.Locked {
		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrThreadAlreadyLocked.Error())
		return
	}
	thread.Locked = true

	log := &data.ThreadLog{
		Action: data.ActionLocked,
		By:     1, // to change
		At:     time.Now().UTC(),
	}

	err = app.models.Messaging.UpdateThread(thread)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Messaging.InsertThreadLog(log)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"thread": thread})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) unlockThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if threadID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchThread.Error())
		return
	}

	thread, err := app.models.Messaging.GetThreadByID(threadID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchThread):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if !thread.Locked {
		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrThreadAlreadyUnlocked.Error())
		return
	}
	thread.Locked = false

	log := &data.ThreadLog{
		Action: data.ActionUnlocked,
		By:     1, // to change
		At:     time.Now().UTC(),
	}

	err = app.models.Messaging.UpdateThread(thread)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Messaging.InsertThreadLog(log)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"thread": thread})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addNewUsersToThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if threadID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchThread.Error())
		return
	}

	thread, err := app.models.Messaging.GetThreadByID(threadID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchThread):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		UserIDs []int `json:"user_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	allUserIDs, err := app.models.Users.GetAllUserIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	badIDs := helpers.VerifyExistsInSlice(input.UserIDs, allUserIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchUsers.Error(), badIDs))
		return
	}

	var addedUsers []int

	for _, id := range input.UserIDs {
		err = app.models.Messaging.AddUserToThread(id, thread.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrUserAlreadyInThread):
				continue
			default:
				app.writeInternalServerError(w, r, err)
				return
			}
		}
		addedUsers = append(addedUsers, id)
	}
	if len(addedUsers) > 0 {
		log := &data.ThreadLog{
			Action:  data.ActionAddedUser,
			Targets: addedUsers,
			By:      1, // to change
			At:      time.Now().UTC(),
		}
		err = app.models.Messaging.InsertThreadLog(log)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	ids, err := app.models.Messaging.GetUserIDsForThread(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"user_ids": ids})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeUsersFromThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if threadID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchThread.Error())
		return
	}

	thread, err := app.models.Messaging.GetThreadByID(threadID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchThread):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		UserIDs []int `json:"user_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	allThreadUserIDs, err := app.models.Messaging.GetUserIDsForThread(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	badIDs := helpers.VerifyExistsInSlice(input.UserIDs, allThreadUserIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrUsersNotInThread.Error(), badIDs))
		return
	}

	var removedUsers []int

	for _, id := range input.UserIDs {
		err = app.models.Messaging.RemoveUserFromThread(id, thread.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		removedUsers = append(removedUsers, id)
	}
	if len(removedUsers) > 0 {
		log := &data.ThreadLog{
			Action:  data.ActionRemovedUser,
			Targets: removedUsers,
			By:      1, // to change
			At:      time.Now().UTC(),
		}
		err = app.models.Messaging.InsertThreadLog(log)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	ids, err := app.models.Messaging.GetUserIDsForThread(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"user_ids": ids})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
