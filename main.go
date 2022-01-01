package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func GitHandlerInfo(w http.ResponseWriter, r *http.Request) {
	user, repo := mux.Vars(r)["user"], mux.Vars(r)["repo"]
	r.URL.Path = "/" + user + "/" + repo + "/info/refs"
	GitHandler.ServeHTTP(w, r)
}

func GitHandlerPackage(w http.ResponseWriter, r *http.Request, pkg string) {
	r.ParseForm()
	user, repo := mux.Vars(r)["user"], mux.Vars(r)["repo"]
	r.URL.Path = "/" + user + "/" + repo + "/" + pkg
	GitHandler.ServeHTTP(w, r)
}

func main() {
	r := mux.NewRouter()

	//handle git operations
	r.HandleFunc(GitRepoInfo, GitHandlerInfo)
	r.HandleFunc(GitRepoRecivePack, func(w http.ResponseWriter, r *http.Request) {
		GitHandlerPackage(w, r, "git-receive-pack")
	})
	r.HandleFunc(GitRepoUploadPack, func(w http.ResponseWriter, r *http.Request) {
		GitHandlerPackage(w, r, "git-upload-pack")
	})

	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
}
