.PHONY: build
build: wasm_exec.js
	yarn install
	go install -v ..
	go-pry -generate="main.go" -i="fmt,log,math" -e='log.Println("Hello world!")'
	GOOS=js GOARCH=wasm go build -v -ldflags "-s -w" -o main.wasm main.go
	go-pry -generate="stdlib.go" -i="$(shell tr '\n' ',' < packages.txt)" -e='log.Println("Hello world!")'
	GOOS=js GOARCH=wasm go build -v -ldflags "-s -w" -o stdlib.wasm stdlib.go

wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	cp $< .

.PHONY: deploy
deploy: build
	now
	now alias
