package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc(GitRepoRecivePack, GitHandler.ServeHTTP)
	r.HandleFunc(GitRepoInfo, GitHandler.ServeHTTP)
	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
	log.Println("partito")
}
