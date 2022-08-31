package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
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

func (app *application) inputJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	var max int64 = 1048576 // 1 MiB
	r.Body = http.MaxBytesReader(w, r.Body, max)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(destination)
	if err != nil {
		var jsonSyntaxError *json.SyntaxError
		var jsonUnmarshalTypeError *json.UnmarshalTypeError
		var jsonInvalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.Is(err, io.EOF):
			return errors.New("empty JSON provided")
		case errors.As(err, &jsonSyntaxError) || errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("malformed JSON provided")
		case errors.As(err, &jsonUnmarshalTypeError):
			return fmt.Errorf("value of invalid type provided for %s", jsonUnmarshalTypeError.Field)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return fmt.Errorf("invalid field %s", strings.TrimPrefix(err.Error(), "json: unknown field "))
		case errors.As(err, &jsonInvalidUnmarshalError):
			panic(err)
		case errors.As(err, new(*http.MaxBytesError)):
			return errors.New("maximum body size is 1 MiB")
		default:
			return err
		}
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("must contain only one JSON object")
	}

	return nil
}

func (app *application) getIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	return ip, nil
}
