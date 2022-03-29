package main

import (
	"log"
	"net/http"

	"snapp/handlers"
	"snapp/limiters"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Root)
	log.Fatal(http.ListenAndServe(":8080", limiters.ByIp(mux, 1, 3)))
}
