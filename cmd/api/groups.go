package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
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
	groups, err := app.models.Groups.GetAllGroups()
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

	err = app.outputJSON(w, http.StatusCreated, envelope{"group": group})
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

	err = app.outputJSON(w, http.StatusOK, envelope{"group": group})
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
		UserIDs []int `json:"user_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	for _, id := range input.UserIDs {
		user, err := app.models.Users.GetUserByID(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoSuchUser):
				app.writeErrorResponse(w, r, http.StatusNotFound, fmt.Sprintf("%s: id %d", err.Error(), id))
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}

		err = app.models.Groups.InsertUserIntoGroup(user.ID, group.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrUserAlreadyInGroup):
				app.writeErrorResponse(w, r, http.StatusConflict, fmt.Sprintf("%s: id %d", err.Error(), user.ID))
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

	for _, id := range input.UserIDs {
		user, err := app.models.Users.GetUserByID(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoSuchUser):
				app.writeErrorResponse(w, r, http.StatusNotFound, fmt.Sprintf("%s: id %d", err.Error(), id))
			default:
				app.writeInternalServerError(w, r, err)
			}
			return
		}

		err = app.models.Groups.RemoveUserFromGroup(user.ID, group.ID)
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
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
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

	err = app.outputJSON(w, http.StatusOK, envelope{"users": users})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
