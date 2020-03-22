# goproxie

[Go](https://golang.org/) clone of [proxie cli](https://github.com/AckeeCZ/be-scripts#proxie) for Node.js.

## User manual

- Use default `goproxie` to start interactive wizard
- Use `goproxie history` to pick a used proxy settings
- Use `goproxie -project=... -cluster=...` for non-interactive mode, see `--help` for all the options available

## Installation

1. Make sure you have Go installed.
2. Get goproxie:
```
go get -u github.com/AckeeCZ/goproxie
```
This will download the source and compile the executable `$GOPATH/bin/goproxie`. Make sure `$GOPATH/bin` is in your `$PATH`.

## Test

Run all tests
```sh
go test ./...
```

See coverage
```sh
go test ./... -v -coverprofile=coverage.out && go tool cover -html=coverage.out
```