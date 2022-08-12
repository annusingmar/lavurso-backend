package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) getGroup(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if groupID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGroup.Error())
		return
	}

	group, err := app.models.Groups.GetGroupByID(groupID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGroup):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	ok, err := app.models.Groups.IsUserInGroup(sessionUser.ID, group.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

	if !ok && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"group": group})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getAllGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := app.models.Groups.GetAllGroups()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	for _, group := range groups {
		count, err := app.models.Groups.GetUserCountForGroup(group.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		group.MemberCount = count
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"groups": groups})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createGroup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	group := &data.Group{
		Name: input.Name,
	}

	v := validator.NewValidator()

	v.Check(group.Name != "", "name", "must be provided")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.models.Groups.InsertGroup(group)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) updateGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if groupID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGroup.Error())
		return
	}

	group, err := app.models.Groups.GetGroupByID(groupID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGroup):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		Name *string `json:"name"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	if input.Name != nil {
		v.Check(*input.Name != "", "name", "must be provided")
		if !v.Valid() {
			app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
			return
		}
		group.Name = *input.Name
	}

	err = app.models.Groups.UpdateGroup(group)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if groupID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGroup.Error())
		return
	}

	group, err := app.models.Groups.GetGroupByID(groupID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGroup):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Groups.DeleteGroup(group.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addUsersToGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if groupID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGroup.Error())
		return
	}

	group, err := app.models.Groups.GetGroupByID(groupID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGroup):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		UserIDs  []int    `json:"user_ids"`
		Roles    []string `json:"roles"`
		ClassIDs []int    `json:"class_ids"`
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

	allClassIDs, err := app.models.Classes.GetAllClassIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs = helpers.VerifyExistsInSlice(input.ClassIDs, allClassIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchClass.Error(), badIDs))
		return
	}

	for _, role := range input.Roles {
		if role != data.RoleAdministrator && role != data.RoleTeacher && role != data.RoleParent && role != data.RoleStudent {
			app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("no such role: %s", role))
			return
		}
	}

	for _, id := range input.UserIDs {
		err = app.models.Groups.InsertUserIntoGroup(id, group.ID)
		if err != nil && !errors.Is(err, data.ErrUserAlreadyInGroup) {
			app.writeInternalServerError(w, r, err)
			return
		}
		err = app.models.Messaging.AddUserGroupToAllThreads(group.ID, id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.ClassIDs {
		users, err := app.models.Classes.GetUsersForClassID(id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		for _, u := range users {
			err = app.models.Groups.InsertUserIntoGroup(u.ID, group.ID)
			if err != nil && !errors.Is(err, data.ErrUserAlreadyInGroup) {
				app.writeInternalServerError(w, r, err)
				return
			}
			err = app.models.Messaging.AddUserGroupToAllThreads(group.ID, u.ID)
			if err != nil {
				app.writeInternalServerError(w, r, err)
				return
			}
		}
	}

	for _, role := range input.Roles {
		users, err := app.models.Users.GetUsersByRole(role)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		for _, u := range users {
			err = app.models.Groups.InsertUserIntoGroup(u.ID, group.ID)
			if err != nil && !errors.Is(err, data.ErrUserAlreadyInGroup) {
				app.writeInternalServerError(w, r, err)
				return
			}
			err = app.models.Messaging.AddUserGroupToAllThreads(group.ID, u.ID)
			if err != nil {
				app.writeInternalServerError(w, r, err)
				return
			}
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeUsersFromGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if groupID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGroup.Error())
		return
	}

	group, err := app.models.Groups.GetGroupByID(groupID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGroup):
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

	for _, id := range input.UserIDs {
		err = app.models.Groups.RemoveUserFromGroup(id, group.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		err = app.models.Messaging.RemoveUserGroupFromAllThreads(group.ID, id)
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

func (app *application) getGroupsForUser(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	switch sessionUser.Role {
	case data.RoleAdministrator:
	default:
		if sessionUser.ID != userID {
			app.notAllowed(w, r)
			return
		}
	}

	user, err := app.models.Users.GetUserByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	groups, err := app.models.Groups.GetGroupsByUserID(user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"groups": groups})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getUsersForGroup(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	groupID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if groupID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGroup.Error())
		return
	}

	group, err := app.models.Groups.GetGroupByID(groupID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGroup):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	switch sessionUser.Role {
	case data.RoleAdministrator:
	default:
		ok, err := app.models.Groups.IsUserInGroup(sessionUser.ID, group.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	users, err := app.models.Groups.GetUsersByGroupID(group.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"group": group, "users": users})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
