package main

import (
	"fmt"
	"os/exec"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u User) Register() error {
	cmd := exec.Command(conf.Scripts+"/newUser.sh", u.Username, u.Password)

	stdout, err := cmd.Output()

	if err != nil {
		return err
	}

	fmt.Print(string(stdout))
	return nil
}
