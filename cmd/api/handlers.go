package main

import (
	"fmt"
	"net/http"
)

func (app *application) tere(w http.ResponseWriter, r *http.Request) {
	app.infoLogger.Println(r.URL, r.RemoteAddr)
	fmt.Fprintf(w, "Tere, maailm!")
}
