package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	
	//handle git operations
	r.HandleFunc(GitRepoInfo, GitHandler.ServeHTTP)
	r.HandleFunc(GitRepoRecivePack, GitHandler.ServeHTTP)

	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
}
