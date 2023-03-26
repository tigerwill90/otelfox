# [WIP] Otelfox

Otelfox is a middleware for the [Fox HTTP router](https://github.com/tigerwill90/fox) that provides distributed 
tracing using [OpenTelemetry](https://opentelemetry.io/).

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
	"fmt"
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/otelfox"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
)

func main() {
	// Create a new Fox router and enable the option to save the matched route.
	r := fox.New(fox.WithSaveMatchedRoute(true))

	tracer := otelfox.New("my-service")

	// Wrap a fox.Handler with the middleware.
	handler := tracer.Middleware(fox.HandlerFunc(func(w http.ResponseWriter, r *http.Request, params fox.Params) {
		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attribute.String("name", params.Get("name")))
		_, _ = fmt.Fprintf(w, "Hello, %s\n", params.Get("name"))
	}))

	// Register the wrapped handler.
	_ = r.Handler(http.MethodGet, "/hello/:name", handler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
```` 

You can also use Otelfox in conjunction with [Foxchain](https://github.com/tigerwill90/foxchain) for seamless integration.

````go
package main

import (
	"fmt"
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/foxchain"
	"github.com/tigerwill90/otelfox"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
)

func main() {
	// Create a new Fox router and enable the option to save the matched route.
	r := fox.New(fox.WithSaveMatchedRoute(true))

	// Create a middleware chain for otelfox.
	chain := foxchain.New(otelfox.New("my-service"))

	// Register the wrapped handler.
	_ = r.Handler(http.MethodGet, "/hello/:name", chain.Then(fox.HandlerFunc(func(w http.ResponseWriter, r *http.Request, params fox.Params) {
		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attribute.String("name", params.Get("name")))
		_, _ = fmt.Fprintf(w, "Hello, %s\n", params.Get("name"))
	})))

	log.Fatal(http.ListenAndServe(":8080", r))
}
````