# goproxie

[![Build Status](https://travis-ci.com/AckeeCZ/goproxie.svg?branch=master)](https://travis-ci.com/AckeeCZ/goproxie)
[![Coverage Status](https://coveralls.io/repos/github/AckeeCZ/goproxie/badge.svg?branch=master)](https://coveralls.io/github/AckeeCZ/goproxie?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/AckeeCZ/goproxie)](https://goreportcard.com/badge/github.com/AckeeCZ/goproxie)


[Go](https://golang.org/) clone of [proxie cli](https://github.com/AckeeCZ/be-scripts#proxie) for Node.js.

## User manual

- Use default `goproxie` to start interactive wizard
- Use `goproxie history` to pick a used proxy settings
- Use `goproxie -project=... -cluster=...` for non-interactive mode, see `--help` for all the options available

## Installation

1. Make sure you have Go, `kubectl` and `gcloud` installed.
2. Authorize `gcloud` to access the Cloud Platform with Google user credentials:
```
gcloud auth login
```
3. Get goproxie:
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
