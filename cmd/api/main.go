package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

// application sisaldab asju, mida tahame terve programmiga jagada,
// selle kylge paneme ka erinevaid funktsioone (reciever function'ite kaudu)
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

	pool := openDBConnection(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName))
	models := data.NewModel(pool)

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

	app.infoLogger.Printf("starting http server on %s", app.config.Web.Listen)
	err := server.ListenAndServe()
	log.Fatalln(err)
}
