// +build !js

package pry

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"

	homedir "github.com/mitchellh/go-homedir"
)

var readFile = ioutil.ReadFile

var historyFile = ".go-pry_history"

func historyPath() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return path.Join(dir, historyFile), nil
}

func loadHistory() []string {
	path, err := historyPath()
	if err != nil {
		log.Printf("Error finding user home dir: %s", err)
		return nil
	}
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var history []string
	if err := json.Unmarshal(body, &history); err != nil {
		log.Printf("Error reading history file! %s", err)
		return nil
	}
	return history
}

func saveHistory(history *[]string) {
	body, err := json.Marshal(history)
	if err != nil {
		log.Printf("Err marshalling history: %s", err)
	}
	path, err := historyPath()
	if err != nil {
		log.Printf("Error finding user home dir: %s", err)
		return
	}
	if err := ioutil.WriteFile(path, body, 0755); err != nil {
		log.Printf("Error writing history: %s", err)
	}
}
