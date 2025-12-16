package main

import (
	"chacalc/src/auth/config"
	"chacalc/src/auth/router"
	"log"
	"net/http"
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
