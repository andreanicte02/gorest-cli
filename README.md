# gorest-cli

[![CI](https://github.com/andreanicte02/gorest-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/andreanicte02/gorest-cli/actions/workflows/ci.yml)
[![Live demo](https://img.shields.io/badge/demo-live-00758f)](https://andreanicte02.github.io/gorest-cli/)

A small, dependency-free Go client and CLI for the [GoRest](https://gorest.co.in) users API —
also compiled to **WebAssembly** so you can try it in the browser.

**▶ Try it: <https://andreanicte02.github.io/gorest-cli/>**

## Features

- `gorest` package: typed models, `context` support, timeouts and structured API errors (including 422 field validation).
- CLI with `list`, `get`, `create`, `update` and `delete` subcommands; JSON output that pipes nicely into `jq`.
- Same package runs in the browser through `syscall/js` — no backend, no JavaScript rewrite.
- Unit tests against `httptest`, CI on every push, cross-platform binaries released with GoReleaser.

## Install

```sh
go install github.com/andreanicte02/gorest-cli/cmd/gorest@latest
```

Or grab a binary for Linux, macOS or Windows from the [releases page](https://github.com/andreanicte02/gorest-cli/releases/latest).

## Usage

Reading is public. Writing needs a free token from <https://gorest.co.in/my-account/access-tokens>:

```sh
export GOREST_TOKEN=your-token
```

```sh
gorest list --page 1 --per-page 5
gorest get 8660809

gorest create --name "Ada Lovelace" --email ada@example.com --gender female
gorest update 8660809 --status inactive
gorest delete 8660809

gorest list | jq '.[].email'
```

As a library:

```go
c := gorest.New(os.Getenv("GOREST_TOKEN"))
user, err := c.CreateUser(ctx, gorest.User{Name: "Ada", Email: "ada@example.com", Gender: "female", Status: "active"})

var apiErr *gorest.APIError
if errors.As(err, &apiErr) && apiErr.StatusCode == 422 {
	for _, f := range apiErr.Fields {
		fmt.Println(f.Field, f.Message)
	}
}
```

## Project layout

```
gorest/       API client package + tests
cmd/gorest/   CLI
cmd/wasm/     WebAssembly bindings for the browser demo
web/          Static demo page (deployed to GitHub Pages)
```

## Development

```sh
make test    # go vet + go test -race
make build   # ./bin/gorest
make serve   # build the wasm demo and serve it on http://localhost:8080
```

Releases are cut by pushing a tag: `git tag v1.0.0 && git push --tags`.

## Background

This started in 2021 as a university assignment (*Software Avanzado*, USAC): a menu-driven
console program that exercised POST/GET/PUT/DELETE against GoRest. It was later rewritten
into an idiomatic, tested Go package with a proper CLI and a WebAssembly demo.

## License

[MIT](LICENSE)
