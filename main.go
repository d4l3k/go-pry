package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/d4l3k/go-pry/generate"
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	// Catch Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
		}
	}()

	if err := run(); err != nil {
		log.Fatal("%+v", err)
	}
}

func run() error {
	ctx := context.Background()

	// FLAGS
	imports := flag.String("i", "fmt,math", "packages to import, comma seperated")
	revert := flag.Bool("r", true, "whether to revert changes on exit")
	execute := flag.String("e", "", "statements to execute")
	generatePath := flag.String("generate", "", "the path to generate a go-pry injected file - EXPERIMENTAL")
	debug := flag.Bool("d", false, "display debug statements")

	flag.CommandLine.Usage = func() {
		if err := generate.NewGenerator(*debug).ExecuteGoCmd(ctx, []string{}, nil); err != nil {
			log.Fatal(err)
		}
		fmt.Println("----")
		fmt.Println("go-pry is an interactive REPL and wrapper around the go command.")
		fmt.Println("You can execute go commands as normal and go-pry will take care of generating the pry code.")
		fmt.Println("Running go-pry with no arguments will drop you into an interactive REPL.")
		flag.PrintDefaults()
		fmt.Println("  revert: cleans up go-pry generated files if not automatically done")
	}
	flag.Parse()

	g := generate.NewGenerator(*debug)

	cmdArgs := flag.Args()
	if len(cmdArgs) == 0 {
		imports := strings.Split(*imports, ",")
		if len(*generatePath) > 0 {
			return g.GenerateFile(imports, *execute, *generatePath)
		}
		return g.GenerateAndExecuteFile(ctx, imports, *execute)
	}

	goDirs := []string{}
	for _, arg := range cmdArgs {
		if strings.HasSuffix(arg, ".go") {
			goDirs = append(goDirs, filepath.Dir(arg))
		}
	}
	if len(goDirs) == 0 {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		goDirs = []string{dir}
	}

	processedFiles := []string{}
	modifiedFiles := []string{}

	if cmdArgs[0] == "revert" {
		fmt.Println("REVERTING PRY")
		for _, dir := range goDirs {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, ".gopry") {
					processed := false
					for _, file := range processedFiles {
						if file == path {
							processed = true
						}
					}
					if !processed {
						base := filepath.Base(path)
						newPath := filepath.Dir(path) + "/" + base[1:len(base)-3]
						modifiedFiles = append(modifiedFiles, newPath)
					}
				}
				return nil
			})
		}
		return g.RevertPry(modifiedFiles)
	}

	testsRequired := cmdArgs[0] == "test"
	for _, dir := range goDirs {
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !testsRequired && strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") {
				return nil
			}
			for _, file := range processedFiles {
				if file == path {
					return nil
				}
			}
			file, err := g.InjectPry(path)
			if err != nil {
				return err
			}
			if file != "" {
				modifiedFiles = append(modifiedFiles, path)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if cmdArgs[0] == "apply" {
		return nil
	}

	if err := g.ExecuteGoCmd(ctx, cmdArgs, nil); err != nil {
		return err
	}

	if *revert {
		if err := g.RevertPry(modifiedFiles); err != nil {
			return err
		}
	}
	return nil
}
