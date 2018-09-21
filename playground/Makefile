.PHONY: build
build: wasm_exec.js
	yarn install
	go install -v ..
	go-pry -generate="main.go" -d -i="$(shell tr '\n' ',' < packages.txt)" -e='log.Println("Hello world!")'
	GOOS=js GOARCH=wasm go build -v -ldflags "-s -w" -o main.wasm

wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	cp $< .

.PHONY: deploy
deploy: build
	now
	now alias
