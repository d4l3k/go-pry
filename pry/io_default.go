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

type IOHistory struct {
	FileName string
	FilePath string
	Records  []string
}

// NewHistory constructs IOHistory instance
func NewHistory() (*IOHistory, error) {
	h := IOHistory{}
	h.FileName = historyFile

	dir, err := homedir.Dir()
	if err != nil {
		log.Printf("Error finding user home dir: %s", err)
		return nil, err
	}
	h.FilePath = path.Join(dir, h.FileName)

	return &h, nil
}

// Load unmarshal history file into history's records
func (h *IOHistory) Load() error {
	body, err := ioutil.ReadFile(h.FilePath)
	if err != nil {
		log.Printf("History file not found! %s", err)
		return err
	}
	var records []string
	if err := json.Unmarshal(body, &records); err != nil {
		log.Printf("Error reading history file! %s", err)
		return err
	}

	h.Records = records
	return nil
}

// Save saves marshaled history's records into file
func (h IOHistory) Save() error {
	body, err := json.Marshal(h.Records)
	if err != nil {
		log.Printf("Err marshalling history: %s", err)
		return err
	}
	if err := ioutil.WriteFile(h.FilePath, body, 0755); err != nil {
		log.Printf("Error writing history: %s", err)
		return err
	}

	return nil
}

// Len returns amount of records in history
func (h IOHistory) Len() int { return len(h.Records) }

// Add appends record into history's records
func (h *IOHistory) Add(record string) {
	h.Records = append(h.Records, record)
}
