package main

import (
	"io/ioutil"
	"log"

	"github.com/sosedoff/gitkit"
	"gopkg.in/yaml.v2"
)

type config struct {
	Scripts       string `yaml:"scriptFolder"`
	Repos         string `yaml:"reposFolder"`
	Port          string `yaml:"port"`
	Uri           string `yaml:"uri"`
	JwtSecret     string `yaml:"jwtSecret"`
	Email         string `yaml:"email"`
	EmailPassword string `yaml:"email-password"`
}

var (
	conf       config
	GitHandler *gitkit.Server
)

func init() {
	//read and unmarshal the yaml config file
	dat, err := ioutil.ReadFile("configs/config.yaml")
	if err != nil {
		log.Fatalf("error reading the config file: %s", err)
	}

	if err := yaml.Unmarshal([]byte(dat), &conf); err != nil {
		log.Fatalf("error unmarshalling the config file: %v", err)
	}

	// Configure git server. Will create git repos path if it does not exist.
	GitHandler = NewServer(conf.Repos)
	if err := GitHandler.Setup(); err != nil {
		log.Fatalf("error setting up the git handler: %s", err.Error())
	}

	// Connect to the database
	if err := ConnectToDatabaseUsers(); err != nil {
		log.Fatalf("error connecting to the database: %s", err.Error())
	}
}
