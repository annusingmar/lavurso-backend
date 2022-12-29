package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) getAllYears(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var err error
	var years []*data.YearExt

	if *sessionUser.Role == data.RoleAdministrator && r.URL.Query().Get("stats") == "true" {
		years, err = app.models.Years.ListAllYearsWithStats()
	} else {
		years, err = app.models.Years.ListAllYears()

	}

	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getYearsForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	student, err := app.models.Users.GetStudentByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if sessionUser.ID != student.ID && *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOrParentOfStudent(student.ID, sessionUser.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	years, err := app.models.Years.GetYearsForStudent(student.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getYearsForClass(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if classID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(classID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	years, err := app.models.Years.GetYearsForClass(class.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) setYearsForClass(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if classID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchClass.Error())
		return
	}

	class, err := app.models.Classes.GetClassByID(classID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchClass):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input []struct {
		YearID int     `json:"year_id"`
		Name   *string `json:"name"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	var yearIDs []int
	var removedYears []int
	var classesYears []*data.ClassYear

	for _, i := range input {
		i := i
		v.Check(i.YearID > 0, "year_id", "must be valid")
		yearIDs = append(yearIDs, i.YearID)
		if i.Name == nil || *i.Name == "" {
			removedYears = append(removedYears, i.YearID)
		} else {
			classesYears = append(classesYears, &data.ClassYear{
				ClassID:     &class.ID,
				YearID:      &i.YearID,
				DisplayName: i.Name,
			})
		}
	}

	allYearIDs, err := app.models.Years.GetAllYearIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	badIDs := helpers.VerifyExistsInSlice(yearIDs, allYearIDs)
	v.Check(badIDs == nil, "year_id", fmt.Sprintf("invalid year id(s): %v", badIDs))

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	for _, cy := range classesYears {
		err = app.models.Years.InsertYearForClass(cy)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	if removedYears != nil {
		err = app.models.Years.RemoveYearsForClass(class.ID, removedYears)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) newYear(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DisplayName string `json:"display_name"`
		NewClasses  []struct {
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
		} `json:"new_classes"`
		TransferredClasses []struct {
			ClassID     int    `json:"class_id"`
			DisplayName string `json:"display_name"`
		} `json:"transferred_classes"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.DisplayName != "", "display_name", "cannot be empty")

	var classIDs []int

	for _, tc := range input.TransferredClasses {
		classIDs = append(classIDs, tc.ClassID)
		v.Check(tc.DisplayName != "", "display_name", fmt.Sprintf("class id %d name cannot be empty", tc.ClassID))
	}

	for _, nc := range input.NewClasses {
		v.Check(nc.DisplayName != "", "display_name", "new class display name cannot be empty")
		v.Check(nc.Name != "", "name", "new class name cannot be empty")
	}

	allClassIDs, err := app.models.Classes.GetAllClassIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs := helpers.VerifyExistsInSlice(classIDs, allClassIDs)
	v.Check(badIDs == nil, "class_id", fmt.Sprintf("invalid class id(s): %v", badIDs))

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	// var archiveIDs []int

	// for _, id := range allClassIDs {
	// 	if !slices.Contains(classIDs, id) {
	// 		archiveIDs = append(archiveIDs, id)
	// 	}
	// }

	year := data.Year{
		DisplayName: &input.DisplayName,
		Current:     helpers.ToPtr(false),
	}

	err = app.models.Years.InsertYear(&year)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	var classYears []*model.ClassesYears

	for _, nc := range input.NewClasses {
		class := &data.Class{
			Name: &nc.Name,
		}

		err := app.models.Classes.InsertClass(class)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		classYears = append(classYears, &model.ClassesYears{YearID: &year.ID, ClassID: &class.ID, DisplayName: &nc.DisplayName})
	}

	for _, tc := range input.TransferredClasses {
		classYears = append(classYears, &model.ClassesYears{
			DisplayName: &tc.DisplayName,
			ClassID:     &tc.ClassID,
			YearID:      &year.ID,
		})
	}

	for _, cy := range classYears {
		err := app.models.Years.InsertYearForClass(cy)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.models.Years.RemoveCurrentYear()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Years.SetYearAsCurrent(year.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	// for _, id := range archiveIDs {
	// 	err = app.models.Users.ArchiveUsersByClassID(id)
	// 	if err != nil {
	// 		app.writeInternalServerError(w, r, err)
	// 		return
	// 	}
	// }

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
