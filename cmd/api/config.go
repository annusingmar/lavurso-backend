package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type configuration struct {
	Web      web      `toml:"web"`
	Database database `toml:"database"`
}

type web struct {
	Listen             string   `toml:"listen"`
	CORSAllowedOrigins []string `toml:"cors_allowed_origins"`
}

type database struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"dbname"`
}

func parseConfig() configuration {
	// default config
	cfg := configuration{
		web{
			Listen:             "127.0.0.1:8080",
			CORSAllowedOrigins: []string{"http://localhost:9000", "http://127.0.0.1:9000"},
		},
		database{
			Host:     "localhost",
			Port:     5432,
			User:     "username",
			Password: "password",
			Name:     "database_name",
		},
	}

	configData, err := os.ReadFile("config.toml")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("INFO config doesn't exist, default values will be used if environment variables aren't set")
		} else {
			panic(err)
		}
	} else {
		log.Println("INFO config exists, but environment variables will override values set in config")
		err = toml.Unmarshal(configData, &cfg)
		if err != nil {
			panic(err)
		}
	}

	cfg.checkEnvironment()

	return cfg
}

func (cfg *configuration) checkEnvironment() {
	val, ok := os.LookupEnv("WEB_LISTEN")
	if ok {
		log.Println("INFO using environment variable WEB_LISTEN")
		cfg.Web.Listen = val
	}

	val, ok = os.LookupEnv("WEB_CORS_ALLOWED_ORIGINS")
	if ok {
		log.Println("INFO using environment variable WEB_CORS_ALLOWED_ORIGINS")
		list := strings.Split(val, ",")
		for i, v := range list {
			list[i] = strings.TrimSpace(v)
		}
		cfg.Web.CORSAllowedOrigins = list
	}

	val, ok = os.LookupEnv("DATABASE_HOST")
	if ok {
		log.Println("INFO using environment variable DATABASE_HOST")
		cfg.Database.Host = val
	}

	val, ok = os.LookupEnv("DATABASE_PORT")
	if ok {
		log.Println("INFO using environment variable DATABASE_PORT")
		port, err := strconv.Atoi(val)
		if err != nil {
			log.Println("ERROR failed reading environment variable DATABASE_PORT, skipping it")
		}
		cfg.Database.Port = port
	}

	val, ok = os.LookupEnv("DATABASE_USER")
	if ok {
		log.Println("INFO using environment variable DATABASE_USER")
		cfg.Database.User = val
	}

	val, ok = os.LookupEnv("DATABASE_PASSWORD")
	if ok {
		log.Println("INFO using environment variable DATABASE_PASSWORD")
		cfg.Database.Password = val
	}

	val, ok = os.LookupEnv("DATABASE_NAME")
	if ok {
		log.Println("INFO using environment variable DATABASE_NAME")
		cfg.Database.Name = val
	}
}
