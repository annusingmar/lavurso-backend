package main

import "net/http"

func (app *application) writeErrorResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	env := envelope{"error": data}
	err := app.outputJSON(w, status, env)
	if err != nil {
		app.errorLogger.Println(r.Method, r.URL.String(), err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}
}

func (app *application) writeInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.errorLogger.Println(r.Method, r.URL.String(), err)
	app.writeErrorResponse(w, r, http.StatusInternalServerError, "internal server error")
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	app.writeErrorResponse(w, r, http.StatusMethodNotAllowed, "method not allowed")
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.writeErrorResponse(w, r, http.StatusNotFound, "not found")
}
