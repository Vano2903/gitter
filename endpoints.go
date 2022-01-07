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
	GetRepos    string = "/git/api/{user}/repos"
	GetRepoInfo string = "/git/api/{user}/repos/{repo}/get"
	GetRepoFile string = "/git/api/{user}/repos/{repo}/get/{file-hash}"
	AddRepo     string = "/git/api/{user}/add/{repo}"
	DeleteRepo  string = "/git/api/{user}/delete/{repo}"
)
