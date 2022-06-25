package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func (app *application) outputJSON(w http.ResponseWriter, status int, env envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)

	return nil

}
