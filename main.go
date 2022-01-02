package main

import (
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
	emailBody += fmt.Sprintf(`<a href='192.168.1.9:8080/git/api/confirm?token=%s' 
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
	fmt.Println()
}

func ConfirmRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token := r.URL.Query().Get("token")
	fmt.Println(token)
	username, email, err := ParseJWT(token)
	fmt.Println(username, email, err)
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

	fmt.Println(DeleteUser(user.User, user.Pass, false))
	statusCode, err := AddUser(username, email, user.Pass, user.Salt, true)
	if err != nil {
		w.WriteHeader(statusCode) //400 | 406
		w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, err.Error())))
		return
	}

	w.WriteHeader(statusCode) //200
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "msg": "%s"}`, statusCode, "confirmed correctly, you can now login")))
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
	r.HandleFunc(Register, AddUserUnconfirmHandler).Methods("POST")
	r.HandleFunc(ConfirmRegistration, ConfirmRegistrationHandler).Methods("GET")

	fmt.Println(conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
}
