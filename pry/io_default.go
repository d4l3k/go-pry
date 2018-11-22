// +build !js

package pry

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var readFile = ioutil.ReadFile

var historyFile = ".go-pry_history"

type ioHistory struct {
	FileName string
	FilePath string
	Records  []string
}

// NewHistory constructs ioHistory instance
func NewHistory() (*ioHistory, error) {
	h := ioHistory{}
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
func (h *ioHistory) Load() error {
	body, err := ioutil.ReadFile(h.FilePath)
	if err != nil {
		return errors.Wrapf(err, "History file not found")
	}
	var records []string
	if err := json.Unmarshal(body, &records); err != nil {
		return errors.Wrapf(err, "Error reading history file")
	}

	h.Records = records
	return nil
}

// Save saves marshaled history's records into file
func (h ioHistory) Save() error {
	body, err := json.Marshal(h.Records)
	if err != nil {
		return errors.Wrapf(err, "error marshaling history")
	}
	if err := ioutil.WriteFile(h.FilePath, body, 0755); err != nil {
		return errors.Wrapf(err, "error writing history to the file")
	}

	return nil
}

// Len returns amount of records in history
func (h ioHistory) Len() int { return len(h.Records) }

// Add appends record into history's records
func (h *ioHistory) Add(record string) {
	h.Records = append(h.Records, record)
}
