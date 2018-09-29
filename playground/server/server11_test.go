// +build go1.11

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGenerateBundle(t *testing.T) {
	t.Parallel()

	remove := func() {
		if err := os.RemoveAll(bundlesDir); err != nil {
			t.Fatal(err)
		}
	}

	remove()

	if err := os.MkdirAll(bundlesDir, 0755); err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodGet, "/wasm/math,fmt", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	if err := generateBundle(resp, r, "math,fmt"); err != nil {
		t.Fatal(err)
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("expected StatusOK got %+v", resp.Code)
	}

	remove()
}
