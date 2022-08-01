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
	"golang.org/x/exp/slices"
)

func (app *application) verifyUserAndGroupIDs(userIDs, groupIDs []int) ([]int, error) {
	if len(userIDs) > 0 {
		allUserIDs, err := app.models.Users.GetAllUserIDs()
		if err != nil {
			return nil, err
		}

		badUserIDs := helpers.VerifyExistsInSlice(userIDs, allUserIDs)
		if badUserIDs != nil {
			return badUserIDs, data.ErrNoSuchUsers
		}
	}

	if len(groupIDs) > 0 {
		allGroupIDs, err := app.models.Groups.GetAllGroupIDs()
		if err != nil {
			return nil, err
		}

		badGroupIDs := helpers.VerifyExistsInSlice(groupIDs, allGroupIDs)
		if badGroupIDs != nil {
			return badGroupIDs, data.ErrNoSuchGroups
		}
	}

	return nil, nil
}

func (app *application) createThread(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		UserIDs  []int  `json:"user_ids"`
		GroupIDs []int  `json:"group_ids"`
	}

	err := app.inputJSON(w, r, &input)
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

	thread := &data.Thread{
		User:      &data.User{ID: sessionUser.ID},
		Title:     input.Title,
		Locked:    false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if !slices.Contains(input.UserIDs, sessionUser.ID) {
		input.UserIDs = append(input.UserIDs, sessionUser.ID)
	}

	badIDs, err := app.verifyUserAndGroupIDs(input.UserIDs, input.GroupIDs)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUsers) || errors.Is(err, data.ErrNoSuchGroups):
			app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", err.Error(), badIDs))
			return
		default:
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.models.Messaging.InsertThread(thread)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	threadMessage := &data.Message{
		ThreadID:  thread.ID,
		User:      &data.User{ID: sessionUser.ID},
		Body:      input.Body,
		Type:      data.MsgTypeThreadStart,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   1,
	}

	err = app.models.Messaging.InsertMessage(threadMessage)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	for _, id := range input.UserIDs {
		err = app.models.Messaging.AddUserToThread(thread.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.GroupIDs {
		err = app.models.Messaging.AddGroupToThread(thread.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"thread": thread})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) deleteThread(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	if thread.User.ID != sessionUser.ID {
		app.notAllowed(w, r)
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
	sessionUser := app.getUserFromContext(r)

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

	if thread.User.ID != sessionUser.ID {
		app.notAllowed(w, r)
		return
	}

	if thread.Locked {
		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrThreadAlreadyLocked.Error())
		return
	}

	err = app.models.Messaging.SetThreadLocked(thread.ID, true)
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
	sessionUser := app.getUserFromContext(r)

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

	if thread.User.ID != sessionUser.ID {
		app.notAllowed(w, r)
		return
	}

	if !thread.Locked {
		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrThreadAlreadyUnlocked.Error())
		return
	}

	err = app.models.Messaging.SetThreadLocked(thread.ID, false)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"thread": thread})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addNewMembersToThread(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	if thread.User.ID != sessionUser.ID {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		UserIDs  []int `json:"user_ids"`
		GroupIDs []int `json:"group_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	badIDs, err := app.verifyUserAndGroupIDs(input.UserIDs, input.GroupIDs)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUsers) || errors.Is(err, data.ErrNoSuchGroups):
			app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", err.Error(), badIDs))
			return
		default:
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.UserIDs {
		err = app.models.Messaging.AddUserToThread(thread.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.GroupIDs {
		err = app.models.Messaging.AddGroupToThread(thread.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeMembersFromThread(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	if thread.User.ID != sessionUser.ID {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		UserIDs  []int `json:"user_ids"`
		GroupIDs []int `json:"group_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	badIDs, err := app.verifyUserAndGroupIDs(input.UserIDs, input.GroupIDs)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUsers) || errors.Is(err, data.ErrNoSuchGroups):
			app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", err.Error(), badIDs))
			return
		default:
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.UserIDs {
		if id == sessionUser.ID {
			continue
		}
		err = app.models.Messaging.RemoveUserFromThread(thread.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.GroupIDs {
		err = app.models.Messaging.RemoveGroupFromThread(thread.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getThreadsForUser(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	if sessionUser.ID != userID {
		app.notAllowed(w, r)
		return
	}

	threads, err := app.models.Messaging.GetThreadsForUser(sessionUser.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	for _, thread := range threads {
		count, err := app.models.Messaging.GetMessageCountForThread(thread.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		thread.MessageCount = &count
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"threads": threads})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createMessage(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	ok, err := app.models.Messaging.IsUserInThread(sessionUser.ID, thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		Body string `json:"body"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Body != "", "body", "must be provided")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	message := &data.Message{
		ThreadID:  thread.ID,
		User:      &data.User{ID: sessionUser.ID, Name: sessionUser.Name, Role: sessionUser.Role},
		Body:      input.Body,
		Type:      data.MsgTypeNormal,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   1,
	}

	err = app.models.Messaging.InsertMessage(message)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Messaging.SetThreadUpdatedAt(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Messaging.SetThreadAsUnreadForAll(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Messaging.SetThreadAsReadForUser(thread.ID, sessionUser.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) updateMessage(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	messageID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if messageID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchMessage.Error())
		return
	}

	message, err := app.models.Messaging.GetMessageByID(messageID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchMessage):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	ok, err := app.models.Messaging.IsUserInThread(sessionUser.ID, message.ThreadID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	if message.User.ID != sessionUser.ID || !ok {
		app.notAllowed(w, r)
		return
	}

	var input struct {
		Body string `json:"body"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.Body != "", "body", "must be provided")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	if message.Body != input.Body {
		message.Body = input.Body
		message.UpdatedAt = time.Now().UTC()

		err = app.models.Messaging.UpdateMessage(message)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.writeErrorResponse(w, r, http.StatusConflict, err.Error())
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) deleteMessage(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	messageID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if messageID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchMessage.Error())
		return
	}

	message, err := app.models.Messaging.GetMessageByID(messageID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchMessage):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	ok, err := app.models.Messaging.IsUserInThread(sessionUser.ID, message.ThreadID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	if message.User.ID != sessionUser.ID || !ok {
		app.notAllowed(w, r)
		return
	}

	if message.Type == data.MsgTypeThreadStart {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrCantDeleteFirstMessage.Error())
		return
	}

	err = app.models.Messaging.DeleteMessage(message.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getThread(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	ok, err := app.models.Messaging.IsUserInThread(sessionUser.ID, thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	messages, err := app.models.Messaging.GetAllMessagesByThreadID(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Messaging.SetThreadAsReadForUser(thread.ID, sessionUser.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"thread": thread, "messages": messages})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getThreadMembers(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

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

	ok, err := app.models.Messaging.IsUserInThread(sessionUser.ID, thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	if !ok {
		app.notAllowed(w, r)
		return
	}

	users, err := app.models.Messaging.GetUsersInThread(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	groups, err := app.models.Messaging.GetGroupsInThread(thread.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"thread": thread, "users": users, "groups": groups})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
