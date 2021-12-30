package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	Scripts string `yaml:"scriptFolder"`
}

var (
	conf config
)

func init() {
	ScriptFolder := os.Getenv("scriptFolder")
	if ScriptFolder == "" {
		dat, err := ioutil.ReadFile("config.yaml")
		err = yaml.Unmarshal([]byte(dat), &conf)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	} else {
		conf.Scripts = ScriptFolder
	}
}
