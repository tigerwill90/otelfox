# [WIP] Otelfox

Otelfox is a middleware for the [Fox](https://github.com/tigerwill90/fox) that provides distributed 
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
	otel := otelfox.New("fox")
	r := fox.New(
		fox.WithMiddleware(otel.Trace),
		fox.WithRouteNotFound(fox.NotFoundHandler(), otel.Trace),
	)

	r.MustHandle(http.MethodGet, "/hello/{name}", func(c fox.Context) {
		_ = c.String(http.StatusOK, "hello %s\n", c.Param("name"))
	})

	log.Fatalln(http.ListenAndServe(":8080", r))
}
````