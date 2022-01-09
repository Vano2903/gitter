package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Vano2903/gitter/internal/email"
	"github.com/gorilla/mux"
)

type Post struct {
	Username string `json:"username, omitempty"` //username of the user
	Email    string `json:"email, omitempty"`    //email of the user
	Password string `json:"password, omitempty"` //password of the user
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

//send email to confirm registration, add this used to the temporary database
func AddUserUnconfirmHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	statusCode, err := AddUser(post.Username, post.Email, post.Password, "", false)
	if err != nil {
		w.WriteHeader(statusCode) //400 | 406
		w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, err.Error())))
		return
	}

	jwt, err := GenerateJWT(post.Username, post.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error generating JWT: ` + err.Error() + `"}`))
		return
	}

	emailBody := `
	<head>
	<style>
		div {
			background-color: #1e1e1e;
			display: grid;
			padding: 0 1rem 1rem 1rem;
			justify-content: center;
			align-items: center;
			border-radius: .2rem;
		}
		#submit, #submit:visited, #submit:active {
			margin: 1rem auto;
			cursor: pointer;
			font-family: inherit;
			font-size: 1rem;
			border-radius: .2rem;
			padding: 1rem 3rem;
			transition: .2s;
			outline: none;
			height: fit-content;
			background-color: #ffcc80;
			border: none;
			color: #000000;
			text-decoration: none;
		}
	
		#submit:hover {
			background-color: #ca9b52;
		}
	
		h1 {
			margin: 0 auto;
			color: #ffffff;
		}
		p {
			margin-top: 2rem;
			width: 100%;
			color: white;
		}
		#delete, #delete:hover, #delete:visited, #delete:active {
			color: #9c64a6;
			text-decoration: none;
		}
		h2 {
			width: 100%;
			color: #ffffff;
			margin: 0 0 1rem 0;
		}
	</style>
	</head>
	<div>
		<h1>Hi, we are almost done, confirm your registration by clicking the button below</h1>`
	emailBody += fmt.Sprintf(`<a href='http://192.168.1.9:8080/git/api/confirm?token=%s' 
	id='submit'>Confirm your registration</a></div>`, jwt)
	fmt.Println(jwt)
	err = email.SendEmail(conf.Email, conf.EmailPassword, post.Email, "Confirm your registration to gitter", emailBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, http.StatusInternalServerError, "error sending the email: "+err.Error())))
		return
	}

	w.WriteHeader(statusCode) //201
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, "added correctly, check your email to confirm your registration")))
}

//confirm registration, add the user to the actual user's database and create his repository
func ConfirmRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token := r.URL.Query().Get("token")
	// fmt.Println(token)
	username, email, err := ParseJWT(token)
	// fmt.Println(username, email, err)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error parsing token"}`))
		return
	}

	user, err := QueryByEmail(email, false)
	fmt.Println(user, err)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) //400
		w.Write([]byte(`{"code": 404, "msg": "User not found"}`))
		return
	}

	if user.User != username {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "User doesnt match, maybe the jwt is invalid, try to register again"}`))
		return
	}

	DeleteUser(user.User, user.Pass, false)
	statusCode, err := AddUser(username, email, user.Pass, user.Salt, true)
	if err != nil {
		w.WriteHeader(statusCode) //400 | 406
		w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, err.Error())))
		return
	}

	w.WriteHeader(statusCode) //200
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, "confirmed correctly")))
}

//Read as post the email and password and return the user's information
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	user, err := QueryByEmail(post.Email, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) //400
		w.Write([]byte(`{"code": 404, "msg": "User not found"}`))
		return
	}

	hash := fmt.Sprintf("%s", sha256.Sum256([]byte(post.Password+":"+user.Salt)))

	if hash != user.Pass {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Password is wrong"}`))
		return
	}
	//

	jsonUser, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error Marshalling JSON"}`))
		return
	}

	w.Write(jsonUser)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
}

//add a repo and initialize it with git init --bare
func AddRepoHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	user, err := QueryByEmail(post.Email, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) //404
		w.Write([]byte(`{"code": 404, "msg": "User not found"}`))
		return
	}

	repo := mux.Vars(r)["repo"]
	exist, err := user.ExistDir(repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error checking the repository: ` + err.Error() + `"}`))
		return
	}

	if exist {
		w.WriteHeader(http.StatusConflict) //409
		w.Write([]byte(`{"code": 409, "msg": "Repository already exists"}`))
		return
	}

	if err := user.CreateRepo(repo); err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error creating the repository: ` + err.Error() + `"}`))
		return
	}
	w.Write([]byte(`{"code": 201, "msg": "Repository created correctly"}`))
}

//read as a post the user credentials and returns all the user's repos
func GetReposHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	user, err := QueryByEmail(post.Email, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) //404
		w.Write([]byte(`{"code": 404, "msg": "User not found"}`))
		return
	}

	repos, err := user.GetRepos()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error getting the repositories: ` + err.Error() + `"}`))
		return
	}

	reposStruct := struct {
		Code  int      `json:"code"`
		Repos []string `json:"repos"`
	}{
		Code:  http.StatusOK,
		Repos: repos,
	}

	json.NewEncoder(w).Encode(reposStruct)
}

func GetRepoInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	user, err := QueryByEmail(post.Email, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) //404
		w.Write([]byte(`{"code": 404, "msg": "User not found"}`))
		return
	}

	repo := mux.Vars(r)["repo"]
	exist, err := user.ExistDir(repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error checking the repository: ` + err.Error() + `"}`))
		return
	}
	if !exist {
		w.WriteHeader(http.StatusNotFound) //404
		w.Write([]byte(`{"code": 404, "msg": "Repository not found"}`))
		return
	}

	repoInfo, err := user.GetRepoInfo(repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error getting the repository info: ` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(repoInfo)
}

func HashHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error Unmarshalling JSON"}`))
		return
	}

	user, err := QueryByEmail(post.Email, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) //404
		w.Write([]byte(`{"code": 404, "msg": "User not found"}`))
		return
	}

	repo := mux.Vars(r)["repo"]
	exist, err := user.ExistDir(repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		w.Write([]byte(`{"code": 500, "msg": "Error checking the repository: ` + err.Error() + `"}`))
		return
	}
	if !exist {
		w.WriteHeader(http.StatusNotFound) //404
		w.Write([]byte(`{"code": 404, "msg": "Repository not found"}`))
		return
	}

	fileHash := mux.Vars(r)["hash"]
	hashType, err := user.GetHashType(repo, fileHash)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		w.Write([]byte(`{"code": 400, "msg": "Error finding the hash: ` + err.Error() + `, check if the hash is correct"}`))
		return
	}

	switch hashType {
	case "commit":
		commitHash, err := user.GetCommitInfo(repo, fileHash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) //500
			w.Write([]byte(`{"code": 500, "msg": "Error getting the commit hash: ` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(w).Encode(commitHash)
	case "blob":
		w.Header().Set("Content-Type", "text/plain")
		blobHash, err := user.GetBlobInfo(repo, fileHash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) //500
			w.Write([]byte(`{"code": 500, "msg": "Error getting the blob hash: ` + err.Error() + `"}`))
			return
		}
		w.Write([]byte(blobHash))
		// json.NewEncoder(w).Encode(blobHash)
	case "tree":
		treeHash, err := user.GetTreeInfo(repo, fileHash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) //500
			w.Write([]byte(`{"code": 500, "msg": "Error getting the tree hash: ` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(w).Encode(treeHash)
	}
}

//TODO validate the credentials of the user when operating with git
//TODO delete a user

//!
//TODO con get repo ritorna anche il tree non solo l'hash
//!
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
	r.HandleFunc(Login, LoginHandler).Methods("POST")
	r.HandleFunc(Register, AddUserUnconfirmHandler).Methods("POST")
	r.HandleFunc(ConfirmRegistration, ConfirmRegistrationHandler).Methods("GET")
	r.HandleFunc(Singoff, DeleteUserHandler).Methods("POST")

	//repo operations handlers
	r.HandleFunc(AddRepo, AddRepoHandler).Methods("POST")
	r.HandleFunc(GetRepos, GetReposHandler).Methods("POST")
	r.HandleFunc(GetRepoInfo, GetRepoInfoHandler).Methods("POST")
	r.HandleFunc(GetRepoHashInfo, HashHandler).Methods("POST")

	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
}
