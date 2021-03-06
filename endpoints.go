package main

const (
	//git endpoints
	GitRepoRecivePack string = "/git/api/{user}/{repo}/git-receive-pack"
	GitRepoUploadPack string = "/git/api/{user}/{repo}/git-upload-pack"
	GitRepoInfo       string = "/git/api/{user}/{repo}/info/refs"

	//user endpoints
	Register            string = "/git/api/singup"
	ConfirmRegistration string = "/git/api/confirm"
	Login               string = "/git/api/login"
	Singoff             string = "/git/api/singoff"

	//repo endpoints
	GetRepos        string = "/git/api/repos"
	GetRepoInfo     string = "/git/api/repos/{repo}/get"
	GetRepoHashInfo string = "/git/api/repos/{repo}/get/{hash}"
	AddRepo         string = "/git/api/add/{repo}"
	DeleteRepo      string = "/git/api/delete/{repo}"
)
