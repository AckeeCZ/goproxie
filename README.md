## ⚠️ Deprecation Notice

As of 2024-02-13, this repository is **deprecated** and no longer actively maintained. We recommend using official [cloud-sql-proxy](https://github.com/GoogleCloudPlatform/cloud-sql-proxy) as a replacement. 

### Key Points:
- No new features or bug fixes will be implemented.
- Issues and pull requests may not be addressed.
- The repository may be archived in the future.

Thank you for your support!


# goproxie


[![Build Status](https://travis-ci.com/AckeeCZ/goproxie.svg?branch=master)](https://travis-ci.com/AckeeCZ/goproxie)
[![Coverage Status](https://coveralls.io/repos/github/AckeeCZ/goproxie/badge.svg?branch=master)](https://coveralls.io/github/AckeeCZ/goproxie?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/AckeeCZ/goproxie)](https://goreportcard.com/badge/github.com/AckeeCZ/goproxie)


[Go](https://golang.org/) clone of [proxie cli](https://github.com/AckeeCZ/be-scripts#proxie) for Node.js.

## User manual

- `goproxie version` to print the version
- Use default `goproxie` to start interactive wizard
- Use `goproxie history` to pick a used proxy settings
- Use `goproxie use` to interactively select and set your default GCP project. `-project` flag available.
- Use `goproxie -project=... -cluster=...` for non-interactive mode, see `--help` for all the options available

## Installation

1. Make sure you have Go, `kubectl` and `gcloud` installed.
2. Authorize `gcloud` to access the Cloud Platform with Google user credentials:
```
gcloud auth login
```
```
gcloud auth application-default login
```
3. Get goproxie:

    a) Build from source
    ```
    go get -u github.com/AckeeCZ/goproxie
    ```
    This will download the source and compile the executable `$GOPATH/bin/goproxie`. Make sure `$GOPATH/bin` is in your `$PATH`.

    b) Download latest executable for macOS, Windows, Linux [here](https://github.com/AckeeCZ/goproxie/releases)

## Test

Run all tests
```sh
go test ./...
```

See coverage
```sh
go test ./... -v -coverprofile=coverage.out && go tool cover -html=coverage.out
```

## Release a new version

- add a new git version tag , prefixed with `v`, e.g. `v12.34.56`
- set it based on last tag and respect [Semantic Versioning](https://semver.org/)
- goproxie will incorporate this tag in it's `version` command during release step when tags are pushed
