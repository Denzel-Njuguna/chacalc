package main

import (
	"log"
	"net/http"

	"chacalc/internal/config"
	"chacalc/internal/router"
)

func main() {
	cfg := config.Load()
	r := router.New(cfg)

	log.Println("Server Starting in port :8000")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
