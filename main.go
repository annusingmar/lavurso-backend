package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func tere(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Tere, maailm!", time.Now())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", tere)
	err := http.ListenAndServe(":8888", mux)
	log.Fatalln(err)
}
