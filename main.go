package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Handle(GitRepo, GitHandler)
	log.Fatal(http.ListenAndServe(":" + conf.Port, r))
}
