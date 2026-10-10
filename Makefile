GOROOT := $(shell go env GOROOT)
# wasm_exec.js moved to lib/wasm in Go 1.24; fall back to misc/wasm for older toolchains.
WASM_EXEC := $(firstword $(wildcard $(GOROOT)/lib/wasm/wasm_exec.js $(GOROOT)/misc/wasm/wasm_exec.js))

.PHONY: build test wasm serve clean

build:
	go build -o bin/gorest ./cmd/gorest

test:
	go vet ./...
	go test -race ./...

wasm:
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/gorest.wasm ./cmd/wasm
	cp "$(WASM_EXEC)" web/wasm_exec.js

serve: wasm
	@echo "Demo running at http://localhost:8080"
	python3 -m http.server 8080 --directory web

clean:
	rm -rf bin dist web/gorest.wasm web/wasm_exec.js
