package otelfox

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

type config struct {
	provider   trace.TracerProvider
	propagator propagation.TextMapPropagator
	carrier    func(r *http.Request) propagation.TextMapCarrier
}

func defaultConfig() *config {
	return &config{
		provider:   otel.GetTracerProvider(),
		propagator: otel.GetTextMapPropagator(),
		carrier: func(r *http.Request) propagation.TextMapCarrier {
			return propagation.HeaderCarrier(r.Header)
		},
	}
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(c *config) {
		if propagators != nil {
			c.propagator = propagators
		}
	})
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		if provider != nil {
			c.provider = provider
		}
	})
}

// WithTextMapCarrier specify a carrier to use for extracting information from http request.
// If none is specified, propagation.HeaderCarrier is used.
func WithTextMapCarrier(fn func(r *http.Request) propagation.TextMapCarrier) Option {
	return optionFunc(func(c *config) {
		if fn != nil {
			c.carrier = fn
		}
	})
}
