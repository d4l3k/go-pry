// +build !js

package pry

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestHistory(t *testing.T) {
	t.Parallel()

	rand.Seed(time.Now().UnixNano())

	history := &ioHistory{}
	history.FileName = ".go-pry_history_test"
	history.FilePath = filepath.Join(os.TempDir(), history.FileName)

	expected := []string{
		"test",
		fmt.Sprintf("rand: %d", rand.Int63()),
	}
	history.Add(expected[0])
	history.Add(expected[1])

	if err := history.Save(); err != nil {
		t.Error("Failed to save history")
	}

	if err := history.Load(); err != nil {
		t.Error("Failed to load history")
	}

	if !reflect.DeepEqual(expected, history.Records) {
		t.Errorf("history.Load() = %+v; expected %+v", history.Records, expected)
	}

	// delete test history file
	err := os.Remove(history.FilePath)
	if err != nil {
		t.Error(err)
	}
}
