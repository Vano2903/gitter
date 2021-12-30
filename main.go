package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc(GitRepoRecivePack, func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]
		repo := mux.Vars(r)["repo"]

		// user, pass, ok := r.BasicAuth()
		// if !ok {
		// 	// return cred, fmt.Errorf("authentication failed")
		// 	fmt.Fprintf(w, "authentication failed")
		// }
		r.URL.Path = user + "/" + repo
		GitHandler.ServeHTTP(w, r)
	})
	r.HandleFunc(GitRepoRecivePack, func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]
		repo := mux.Vars(r)["repo"]

		// user, pass, ok := r.BasicAuth()
		// if !ok {
		// 	// return cred, fmt.Errorf("authentication failed")
		// 	fmt.Fprintf(w, "authentication failed")
		// }
		r.URL.Path = user + "/" + repo
		GitHandler.ServeHTTP(w, r)
	})
	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
	log.Println("partito")
}
