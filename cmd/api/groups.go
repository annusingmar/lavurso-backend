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

	err = app.outputJSON(w, http.StatusOK, envelope{"group": group})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getAllGroups(w http.ResponseWriter, r *http.Request) {
	var archived bool

	archivedParam := r.URL.Query().Get("archived")

	if archivedParam == "true" {
		archived = true
	} else {
		archived = false
	}

	groups, err := app.models.Groups.GetAllGroups(archived)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
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
		Name: &input.Name,
	}

	v := validator.NewValidator()

	v.Check(*group.Name != "", "name", "must be provided")

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
		Name     *string `json:"name"`
		Archived *bool   `json:"archived"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	if input.Name != nil {
		group.Name = input.Name
	}
	if input.Archived != nil {
		group.Archived = input.Archived
	}

	v.Check(*group.Name != "", "name", "must be provided")
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
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

func (app *application) deleteGroup(w http.ResponseWriter, r *http.Request) {
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

	if *group.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrGroupArchived.Error())
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

	if len(input.UserIDs) > 0 {
		err = app.models.Groups.InsertUsersIntoGroup(input.UserIDs, group.ID)
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

		var ids []int
		for _, u := range users {
			ids = append(ids, u.ID)
		}

		if len(ids) > 0 {
			err = app.models.Groups.InsertUsersIntoGroup(ids, group.ID)
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

		var ids []int
		for _, u := range users {
			ids = append(ids, u.ID)
		}

		if len(ids) > 0 {
			err = app.models.Groups.InsertUsersIntoGroup(ids, group.ID)
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

	if *group.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrGroupArchived.Error())
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

	if len(input.UserIDs) > 0 {
		err = app.models.Groups.RemoveUsersFromGroup(input.UserIDs, group.ID)
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

	switch *sessionUser.Role {
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

	var groups []*data.GroupExt

	if *sessionUser.Role == data.RoleAdministrator || *sessionUser.Role == data.RoleTeacher {
		groups, err = app.models.Groups.GetAllGroups(false)
	} else {
		groups, err = app.models.Groups.GetGroupsByUserID(user.ID)
	}

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
