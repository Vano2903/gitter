package main

const (
	//git endpoints
	GitRepoRecivePack string = "/git/api/{user}/{repo}/git-receive-pack"
	GitRepoUploadPack string = "/git/api/{user}/{repo}/git-upload-pack"
	GitRepoInfo       string = "/git/api/{user}/{repo}/info/refs"

	//user endpoints
	NewUser    string = "/git/api/newUser/{user}"
	DeleteUser string = "/git/api/deleteUser/{user}"

	//repo endpoints
	GetRepos   string = "/git/api/{user}/repos"
	GetRepo    string = "/git/api/{user}/{repo}/get"
	AddRepo    string = "/git/api/{user}/add/{repo}"
	DeleteRepo string = "/git/api/{user}/delete/{repo}"
	//to commit
)
