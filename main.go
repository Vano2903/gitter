package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func GitHandlerFunction(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	repo := mux.Vars(r)["repo"]
	r.URL.Path = user + "/" + repo
	GitHandler.ServeHTTP(w, r)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc(GitRepoRecivePack, GitHandlerFunction)
	r.HandleFunc(GitRepoRecivePack, GitHandlerFunction)
	r.HandleFunc(GitRepo, GitHandlerFunction)
	r.HandleFunc(GitRepo2, GitHandlerFunction)
	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
	log.Println("partito")
}
