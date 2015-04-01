# go-pry

go-pry - an interactive REPL for Go that allows you to drop into your code at any point.

![go-pry](https://i.imgur.com/yr1BEsK.png)




## Usage
Install go-pry
```bash
go get github.com/d4l3k/go-pry
go install github.com/d4l3k/go-pry

```

Add the pry statement to the code
```go
package main

import "github.com/d4l3k/go-pry/pry"

func main() {
  a := 1
  pry.Pry()
}
```

Run the code as you would normally with the `go` command. go-pry is just a wrapper.
```bash
# Run
go-pry run readme.go
```

## How does it work?
go-pry is built using a combination of meta programming as well as a massive amount of reflection. When you invoke the go-pry command it looks at the Go files in the mentioned directories (or the current in cases such as `go-pry build`) and processes them. Since Go is a compiled language there's no way to dynamically get in scope variables, and even if there was, unused imports would be automatically removed for optimization purposes. Thus, go-pry has to find every instance of `pry.Pry()` and inject a large blob of code that contains references to all in scope variables and functions as well as those of the imported packages. When doing this it makes a copy of your file to `.<filename>.gopry` and modifies the `<filename>.go` then passes the command arguments to the standard `go` command. Once the command exits, it restores the files.

If the program unexpectedly fails there is a custom command `go-pry restore [files]` that will move the files back. An alternative is to just remove the `pry.Apply(...)` line.

## Inspiration

go-pry is greatly inspired by [Pry REPL](http://pryrepl.org) for Ruby.

## License

go-pry is licensed under the MIT license.
