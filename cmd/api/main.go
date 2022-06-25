package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/annusingmar/lavurso-backend/internal/data"
)

// configuration sisaldab konfigureeritavaid vaartusi, enamasti command-line options'ite kaudu
type configuration struct {
	listen   string
	database struct {
		connectionString string
	}
}

// application sisaldab asju, mida tahame terve programmiga jagada,
// selle kylge paneme ka erinevaid funktsioone (reciever function'ite kaudu)
type application struct {
	config      configuration
	infoLogger  *log.Logger
	errorLogger *log.Logger
	models      data.Models
}

func main() {
	var config configuration

	flag.StringVar(&config.listen, "listen", "127.0.0.1:8888", "address for HTTP server to listen on, in format ip:port")
	flag.StringVar(&config.database.connectionString, "dsn", "postgres://username:password@localhost:5432/database_name", "connection string for PostgreSQL")
	flag.Parse()

	infoLogger := log.New(os.Stdout, "INFO ", log.Ltime|log.Ldate)
	errorLogger := log.New(os.Stderr, "ERROR ", log.Ltime|log.Ldate)

	pool := openDBConnection(config.database.connectionString)
	models := data.NewModel(pool)

	app := &application{
		config:      config,
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		models:      models,
	}

	server := &http.Server{
		Addr:     app.config.listen,
		ErrorLog: errorLogger,
		Handler:  app.routes(),
	}

	app.infoLogger.Printf("starting http server on %s", config.listen)
	err := server.ListenAndServe()
	log.Fatalln(err)
}
