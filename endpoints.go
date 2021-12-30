package main

const (
	//user endpoints
	NewUser    string = "/api/newUser/:user"
	DeleteUser string = "/api/deleteUser/:user"

	//repo endpoints
	GetRepos   string = "/api/:user/repos"
	GetRepo    string = "/api/:user/:repo"
	AddRepo    string = "/api/:user/add/:repo"
	DeleteRepo string = "/api/:user/delete/:repo"
)
