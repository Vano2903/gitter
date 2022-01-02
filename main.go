package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Post struct {
	Username string `json:"username, omitempty"` //username of the user
	Email    string `json:"email, omitempty"`    //email of the user
	Password string `json:"password, omitempty"` //password of the user
	// Year     int    `json:"year, omitempty"`     //year for the commits
	// Id       string `json:"id, omitempty"`       //id is the id of the game to delete
}

//handle git endpoint /info/refs
func GitHandlerInfo(w http.ResponseWriter, r *http.Request) {
	user, repo := mux.Vars(r)["user"], mux.Vars(r)["repo"]
	r.URL.Path = "/" + user + "/" + repo + "/info/refs"
	GitHandler.ServeHTTP(w, r)
}

//handle git endpoints /git-receive-pack and /git-upload-pack
func GitHandlerPackage(w http.ResponseWriter, r *http.Request, pkg string) {
	r.ParseForm()
	user, repo := mux.Vars(r)["user"], mux.Vars(r)["repo"]
	r.URL.Path = "/" + user + "/" + repo + "/" + pkg
	GitHandler.ServeHTTP(w, r)
}

//add a user to the database and create the user's repo
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	statusCode, err := AddUser(post.Username, post.Email, post.Password)
	if err != nil {
		w.WriteHeader(statusCode) //400 | 406
		w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, err.Error())))
		return
	}
	w.WriteHeader(statusCode) //201
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, "user added successfully")))
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

	//handle user operations
	r.HandleFunc(Register, AddUserHandler).Methods("POST")

	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
}
