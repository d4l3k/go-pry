package pry

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/d4l3k/go-pry/pry/safebuffer"
	"github.com/pkg/errors"
)

func TestCLIBasicStatement(t *testing.T) {
	t.Parallel()

	env := testPryApply(t)
	defer env.Close()

	env.Write([]byte("a := 10\n"))

	succeedsSoon(t, func() error {
		out, _ := env.Get("a")
		want := 10
		if !reflect.DeepEqual(out, want) {
			return errors.Errorf(
				"expected a = %d; got %d\nOutput:\n%s\n", want, out, env.Output())
		}
		return nil
	})
}

func TestCLIHistory(t *testing.T) {
	t.Parallel()

	env := testPryApply(t)
	defer env.Close()

	env.Write([]byte("var a int\na = 1\na = 2\na = 3\n"))
	// down down up up up down enter
	env.Write([]byte("\x1b\x5b\x42\x1b\x5b\x42\x1b\x5b\x41\x1b\x5b\x41\x1b\x5b\x41\x1b\x5b\x42\n"))

	succeedsSoon(t, func() error {
		out, _ := env.Get("a")
		want := 2
		if !reflect.DeepEqual(out, want) {
			return errors.Errorf(
				"expected a = %d; got %d\nOutput:\n%s\n", want, out, env.Output())
		}
		return nil
	})
}

func TestCLIEditingArrows(t *testing.T) {
	t.Parallel()

	env := testPryApply(t)
	defer env.Close()

	env.Write([]byte("a := 100"))
	// left left backspace 2 right right 5 enter
	env.Write([]byte("\x1b\x5b\x44\x1b\x5b\x44\b2\x1b\x5b\x43\x1b\x5b\x435\n"))

	succeedsSoon(t, func() error {
		out, _ := env.Get("a")
		want := 2005
		if !reflect.DeepEqual(out, want) {
			return errors.Errorf(
				"expected a = %d; got %d\nOutput:\n%s\n", want, out, env.Output())
		}
		return nil
	})
}

type testTTY struct {
	*io.PipeReader
	*io.PipeWriter
}

func makeTestTTY() *testTTY {
	r, w := io.Pipe()
	return &testTTY{r, w}
}

func (t *testTTY) ReadRune() (rune, error) {
	buf := make([]byte, 1)
	_, err := t.PipeReader.Read(buf)
	return rune(buf[0]), err
}

func (t *testTTY) Size() (int, int, error) {
	return 10000, 100, nil
}

func (t *testTTY) Close() error {
	t.PipeReader.Close()
	return t.PipeWriter.Close()
}

type testPryEnv struct {
	stdout *safebuffer.Buffer
	*testTTY
	*Scope
	dir, file string
}

func testPryApply(t testing.TB) *testPryEnv {
	var stdout safebuffer.Buffer
	tty := makeTestTTY()
	scope := NewScope()

	dir, err := ioutil.TempDir("", "go-pry-test")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(
		path.Join(dir, "main.go"),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0755,
	)
	if err != nil {
		t.Fatal(err)
	}
	file.Write([]byte(
		`package main

import "github.com/d4l3k/go-pry/pry"

func main() {
	pry.Pry()
}
`,
	))
	file.Close()

	filePath := file.Name()
	lineNum := 2

	go func() {
		if err := apply(scope, &stdout, tty, filePath, filePath, lineNum); err != nil {
			panic(err)
		}
	}()

	return &testPryEnv{
		stdout:  &stdout,
		testTTY: tty,
		Scope:   scope,
		dir:     dir,
		file:    filePath,
	}
}

func (env *testPryEnv) Output() string {
	return env.stdout.String()
}

func (env *testPryEnv) Close() {
	env.Write([]byte("\nexit\n"))
	env.testTTY.Close()
	os.Remove(env.file)
	os.Remove(env.dir)
}

func succeedsSoon(t testing.TB, f func() error) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 5 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		t.Fatal(errors.Wrapf(err, "failed after 5 seconds"))
	}
}
