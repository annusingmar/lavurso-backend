package main

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// configuration sisaldab konfigureeritavaid vaartusi, enamasti config.toml file'i kaudu
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
	DBName   string `toml:"dbname"`
}

// POC, needs to be replaced with a not so bad function
func envCheckVariable(cfg configuration) configuration {

	driver, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		log.Println("WARNING no enviroment variable set for POSTGRES_PASSWORD, reading from configuration file")
	} else {
		cfg.Database.Password = driver
	}
	driver, ok = os.LookupEnv("POSTGRES_DATABASE")
	if !ok {
		log.Println("WARNING no enviroment variable set for POSTGRES_DATABASE, reading from configuration file")
	} else {
		cfg.Database.DBName = driver
	}
	driver, ok = os.LookupEnv("POSTGRES_USER")
	if !ok {
		log.Println("WARNING no enviroment variable set for POSTGRES_USER, reading from configuration file")
	} else {
		cfg.Database.User = driver
	}

	driver, ok = os.LookupEnv("BACKEND_CORS_EXTRA_ALLOWED_ORIGIN")
	if !ok {
		log.Println("WARNING no enviroment variable set for BACKEND_CORS_EXTRA_ALLOWED_ORIGIN, reading from configuration file")
	} else {
		cfg.Web.CORSAllowedOrigins = append(cfg.Web.CORSAllowedOrigins, driver)
	}

	log.Println(cfg)
	return cfg
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
			DBName:   "database_name",
		},
	}

	configData, err := os.ReadFile("config.toml")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("WARNING config doesn't exist, using default values")
			return cfg
		} else {
			panic(err)
		}
	}

	err = toml.Unmarshal(configData, &cfg)
	if err != nil {
		panic(err)
	}

	cfg = envCheckVariable(cfg)

	return cfg
}
