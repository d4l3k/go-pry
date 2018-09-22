.PHONY: build
build: wasm_exec.js
	yarn install
	go install -v ..
	@rm -r bundles
	mkdir -p bundles
	go-pry -generate="bundles/main.go" -i="fmt,log,math" -e='log.Println("Hello world!")'
	GOOS=js GOARCH=wasm go build -v -ldflags "-s -w" -o bundles/main.wasm bundles/main.go
	go-pry -generate="bundles/stdlib.go" -i="$(shell tr '\n' ',' < packages.txt)" -e='log.Println("Hello world!")'
	GOOS=js GOARCH=wasm go build -v -ldflags "-s -w" -o bundles/stdlib.wasm bundles/stdlib.go
	CGO_ENABLED=0 go build -ldflags "-s -w" -o playground ./server

.PHONY: run
run: build
	./playground

wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	cp $< .

.PHONY: deploy
deploy: build
	now
	now alias
