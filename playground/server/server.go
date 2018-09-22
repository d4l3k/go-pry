package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/d4l3k/go-pry/generate"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var bind = flag.String("bind", ":8080", "address to bind to")

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func pkgHash(pkgs []string) string {
	hash := sha256.Sum256([]byte(strings.Join(pkgs, ",")))
	return hex.EncodeToString(hash[:])
}

func normalizePackages(packages string) []string {
	var pkgs []string
	for _, pkg := range strings.Split(strings.ToLower(packages), ",") {
		pkg = strings.TrimSpace(pkg)
		if len(pkg) == 0 {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	return pkgs
}

func generateBundle(w http.ResponseWriter, r *http.Request, packages string) (retErr error) {
	pkgs := normalizePackages(packages)
	hash := pkgHash(pkgs)
	path := filepath.Join("bundles", hash+".wasm")
	goPath := filepath.Join("bundles", hash+".go")
	_, err := os.Stat(path)
	if err == nil {
		http.ServeFile(w, r, path)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	dir, err := ioutil.TempDir("", "pry-playground-gopath")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}()

	env := []string{
		fmt.Sprintf("GOPATH=%s", dir),
	}

	g := generate.NewGenerator(false)
	g.Build.GOPATH = dir
	g.Build.CgoEnabled = false

	for _, pkg := range append([]string{"github.com/d4l3k/go-pry/pry"}, pkgs...) {
		if err := g.ExecuteGoCmd(r.Context(), []string{
			"get",
			pkg,
		}, env); err != nil {
			return errors.Wrapf(err, "error go get %q", pkg)
		}
	}

	if err := g.GenerateFile(pkgs, "", goPath); err != nil {
		return errors.Wrap(err, "GenerateFile")
	}

	if err := g.ExecuteGoCmd(r.Context(), []string{
		"build",
		"-ldflags",
		"-s -w",
		"-o",
		path,
		goPath,
	}, append([]string{
		"GOOS=js",
		"GOARCH=wasm",
	}, env...)); err != nil {
		return errors.Wrapf(err, "go build")
	}

	http.ServeFile(w, r, path)
	return nil
}

func run() error {
	log.SetFlags(log.Flags() | log.Lshortfile)

	router := mux.NewRouter()
	router.PathPrefix("/wasm/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkgs := strings.Join(strings.Split(r.URL.Path, "/")[2:], "/")
		if err := generateBundle(w, r, pkgs); err != nil {
			http.Error(w, fmt.Sprintf("%+v", err), http.StatusInternalServerError)
		}
	})
	router.NotFoundHandler = http.FileServer(http.Dir("."))

	log.Printf("Listening %s...", *bind)
	r := handlers.CombinedLoggingHandler(os.Stderr, router)
	return http.ListenAndServe(*bind, r)
}
