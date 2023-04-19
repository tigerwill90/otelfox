[![Go Reference](https://pkg.go.dev/badge/github.com/tigerwill90/otelfox.svg)](https://pkg.go.dev/github.com/tigerwill90/otelfox)
[![tests](https://github.com/tigerwill90/otelfox/actions/workflows/tests.yaml/badge.svg)](https://github.com/tigerwill90/otelfox/actions?query=workflow%3Atests)
[![Go Report Card](https://goreportcard.com/badge/github.com/tigerwill90/otelfox)](https://goreportcard.com/report/github.com/tigerwill90/otelfox)
[![codecov](https://codecov.io/gh/tigerwill90/otelfox/branch/master/graph/badge.svg?token=D6qSTlzEcE)](https://codecov.io/gh/tigerwill90/otelfox)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tigerwill90/otelfox)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tigerwill90/otelfox)
# Otelfox

Otelfox is a middleware for [Fox](https://github.com/tigerwill90/fox) that provides distributed 
tracing using [OpenTelemetry](https://opentelemetry.io/).

## Disclaimer
Otelfox's API is linked to Fox router, and it will only reach v1 when the router is stabilized.
During the pre-v1 phase, breaking changes may occur and will be documented in the release notes.

## Getting started
### Installation
````shell
go get -u github.com/tigerwill90/otelfox
````

### Features

- Automatically creates spans for incoming HTTP requests
- Extracts and propagates trace context from incoming requests
- Annotates spans with HTTP-specific attributes, such as method, route, and status code

### Usage
````go
package main

import (
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/otelfox"
	"log"
	"net/http"
)

func main() {
	r := fox.New(
		fox.WithMiddleware(otelfox.Middleware("fox")),
	)

	r.MustHandle(http.MethodGet, "/hello/{name}", func(c fox.Context) {
		_ = c.String(http.StatusOK, "hello %s\n", c.Param("name"))
	})

	log.Fatalln(http.ListenAndServe(":8080", r))
}
````
