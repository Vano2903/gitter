package main

import (
	"io/ioutil"
	"log"

	"github.com/sosedoff/gitkit"
	"gopkg.in/yaml.v2"
)

type config struct {
	Scripts string `yaml:"scriptFolder"`
	Repos   string `yaml:"reposFolder"`
	Port string `yaml:"port"`
}

var (
	conf       config
	GitHandler *gitkit.Server
)

func init() {
	// ScriptFolder := os.Getenv("scriptFolder")
	// ScriptFolder := os.Getenv("scriptFolder")
	// if ScriptFolder == "" {
	dat, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal([]byte(dat), &conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// } else {
	// 	conf.Scripts = ScriptFolder
	// }

	GitHandler = NewServer(conf.Repos)
	// Configure git server. Will create git repos path if it does not exist.
	// If hooks are set, it will also update all repos with new version of hook scripts.
	if err := GitHandler.Setup(); err != nil {
		log.Fatal(err)
	}
}
