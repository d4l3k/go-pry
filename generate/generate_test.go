package generate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportPry(t *testing.T) {
	g := NewGenerator(false)
	file := "../example/file/file.go"
	res, err := g.InjectPry(file)

	if err != nil {
		t.Errorf("Failed to inject pry %v", err)
	}

	if !fileExists(res) {
		t.Error("Source file not found")
	}

	pryFile := filepath.Join(filepath.Dir(res), ".file.gopry")
	if !fileExists(pryFile) {
		t.Error("Pry file not found")
	}

	// clean up
	g.RevertPry([]string{res})

	if !fileExists(file) {
		t.Error("Source file not found")
	}

	res, err = g.InjectPry("nonexisting.go")
	if res != "" {
		t.Error("Non empty result received")
	}

	if fileExists(".nonexisting.gopry") {
		t.Error("Pry file should not exists")
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
