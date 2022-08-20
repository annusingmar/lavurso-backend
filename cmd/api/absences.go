package main

// func (app *application) excuseAbsenceForStudent(w http.ResponseWriter, r *http.Request) {
// 	sessionUser := app.getUserFromContext(r)

// 	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
// 	if userID < 0 || err != nil {
// 		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
// 		return
// 	}

// 	switch *sessionUser.Role {
// 	case data.RoleAdministrator:
// 	case data.RoleParent:
// 		ok, err := app.models.Users.IsParentOfChild(*sessionUser.ID, userID)
// 		if err != nil {
// 			app.writeInternalServerError(w, r, err)
// 			return
// 		}
// 		if !ok {
// 			app.notAllowed(w, r)
// 			return
// 		}
// 	case data.RoleTeacher:
// 	default:
// 		app.notAllowed(w, r)
// 		return
// 	}

// 	user, err := app.models.Users.GetUserByID(userID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrNoSuchUser):
// 			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
// 		default:
// 			app.writeInternalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	if *user.Role != data.RoleStudent {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
// 		return
// 	}

// 	var input struct {
// 		MarkID int    `json:"mark_id"`
// 		Excuse string `json:"excuse"`
// 	}

// 	err = app.inputJSON(w, r, &input)
// 	if err != nil {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	at := time.Now().UTC()

// 	excuse := &data.AbsenceExcuse{
// 		MarkID: &input.MarkID,
// 		Excuse: &input.Excuse,
// 		By:     &data.User{ID: sessionUser.ID},
// 		At:     &at,
// 	}

// 	v := validator.NewValidator()

// 	v.Check(*excuse.Excuse != "", "excuse", "must be provided")
// 	v.Check(*excuse.MarkID > 0, "absence_id", "must be provided and valid")

// 	if !v.Valid() {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
// 		return
// 	}

// 	absence, err := app.models.Absences.GetAbsenceByMarkID(*excuse.MarkID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrNoSuchAbsence):
// 			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
// 		default:
// 			app.writeInternalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	if *sessionUser.Role == data.RoleTeacher {
// 		journal, err := app.models.Journals.GetJournalByID(*absence.JournalID)
// 		if err != nil {
// 			app.writeInternalServerError(w, r, err)
// 			return
// 		}
// 		if *sessionUser.ID != *journal.Teacher.ID {
// 			app.notAllowed(w, r)
// 			return
// 		}
// 	}

// 	if absence.UserID != *user.ID {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotValidAbsence.Error())
// 		return
// 	}

// 	if absence.AbsenceExcuse != nil {
// 		app.writeErrorResponse(w, r, http.StatusConflict, data.ErrAbsenceExcused.Error())
// 		return
// 	}

// 	err = app.models.Absences.InsertExcuse(excuse)
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 		return
// 	}

// 	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 		return
// 	}

// }

// func (app *application) deleteAbsenceExcuseForStudent(w http.ResponseWriter, r *http.Request) {
// 	sessionUser := app.getUserFromContext(r)

// 	userID, err := strconv.Atoi(chi.URLParam(r, "sid"))
// 	if userID < 0 || err != nil {
// 		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
// 		return
// 	}

// 	switch *sessionUser.Role {
// 	case data.RoleAdministrator:
// 	case data.RoleParent:
// 		ok, err := app.models.Users.IsParentOfChild(*sessionUser.ID, userID)
// 		if err != nil {
// 			app.writeInternalServerError(w, r, err)
// 			return
// 		}
// 		if !ok {
// 			app.notAllowed(w, r)
// 			return
// 		}
// 	case data.RoleTeacher:
// 	default:
// 		app.notAllowed(w, r)
// 		return
// 	}

// 	user, err := app.models.Users.GetUserByID(userID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrNoSuchUser):
// 			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
// 		default:
// 			app.writeInternalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	if *user.Role != data.RoleStudent {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
// 		return
// 	}

// 	excuseID, err := strconv.Atoi(chi.URLParam(r, "eid"))
// 	if excuseID < 0 || err != nil {
// 		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchExcuse.Error())
// 		return
// 	}

// 	excuse, err := app.models.Absences.GetExcuseByID(excuseID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrNoSuchExcuse):
// 			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
// 		default:
// 			app.writeInternalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	absence, err := app.models.Absences.GetAbsenceByMarkID(*excuse.MarkID)
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 	}

// 	if *sessionUser.Role == data.RoleTeacher {
// 		journal, err := app.models.Journals.GetJournalByID(*absence.JournalID)
// 		if err != nil {
// 			app.writeInternalServerError(w, r, err)
// 			return
// 		}
// 		if *sessionUser.ID != *journal.Teacher.ID {
// 			app.notAllowed(w, r)
// 			return
// 		}
// 	}

// 	if absence.UserID != *user.ID {
// 		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotValidAbsence.Error())
// 		return
// 	}

// 	err = app.models.Absences.DeleteExcuse(*excuse.ID)
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 		return
// 	}

// 	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
// 	if err != nil {
// 		app.writeInternalServerError(w, r, err)
// 	}

// }
