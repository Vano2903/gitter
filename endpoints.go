package main

const (
	//user endpoints
	newUser    string = "/api/newUser/:user"
	deleteUser string = "/api/deleteUser/:user"

	//repo endpoints
	getRepos   string = "/api/:user/repos"
	getRepo    string = "/api/:user/:repo"
	addRepo    string = "api/:user/add/:repo"
	deleteRepo string = "api/:user/delete/:repo"
)
