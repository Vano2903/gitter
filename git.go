package main

import (
	"os"
	"os/exec"

	"github.com/sosedoff/gitkit"
)

func NewServer(repoDir string) *gitkit.Server {
	// Configure git service
	server := gitkit.New(gitkit.Config{
		Dir:        repoDir,
		AutoCreate: true,
	})

	return server
}

func CreateNewDir(path string, gitInit bool) error {
	// create new directory
	if err := os.Mkdir(path, 0755); err != nil {
		return err
	}
	//run git init --bare command
	if gitInit {
		cmd := exec.Command("git", "init", "--bare", "--quiet", path)
		return cmd.Run()
	}
	return nil
}
