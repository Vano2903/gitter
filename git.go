package main

import "github.com/sosedoff/gitkit"

func NewServer(repoDir string) *gitkit.Server {
	// Configure git hooks
	hooks := &gitkit.HookScripts{
		PreReceive: `echo "Hello World!"`,
	}

	// Configure git service
	server := gitkit.New(gitkit.Config{
		Dir:        repoDir,
		AutoCreate: true,
		AutoHooks:  true,
		Hooks:      hooks,
	})

	return server
}