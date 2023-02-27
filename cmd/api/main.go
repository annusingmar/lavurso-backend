package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

type application struct {
	config      configuration
	infoLogger  *log.Logger
	errorLogger *log.Logger
	models      data.Models
}

func main() {
	config := parseConfig()

	infoLogger := log.New(os.Stdout, "INFO ", log.Ltime|log.Ldate)
	errorLogger := log.New(os.Stderr, "ERROR ", log.Ltime|log.Ldate)

	db := config.Database.openConnection()
	models := data.NewModel(db)

	app := &application{
		config:      config,
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		models:      models,
	}

	server := &http.Server{
		Addr:     app.config.Web.Listen,
		ErrorLog: errorLogger,
		Handler:  app.routes(),
	}

	catchSignal := make(chan os.Signal, 1)
	signal.Notify(catchSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		app.infoLogger.Printf("starting HTTP server on %s", app.config.Web.Listen)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.errorLogger.Fatalln(err)
		}
	}()

	<-catchSignal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.infoLogger.Println("shutting down...")
	err := server.Shutdown(ctx)
	if err != nil {
		app.errorLogger.Fatalln(err)
	}
}
