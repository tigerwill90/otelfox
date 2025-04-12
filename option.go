package otelfox

import (
	"github.com/tigerwill90/fox"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

// Filter is a predicate used to determine whether a given http.request should
// be traced. A Filter must return true if the request should be traced.
type Filter func(c fox.Context) bool

// SpanNameFormatter is a function that formats the span name given the HTTP request.
// This allows for dynamic naming of spans based on attributes of the request.
type SpanNameFormatter func(c fox.Context) string

// SpanAttributesFunc is a function type that can be used to dynamically
// generate span attributes for a given HTTP request. It is used in
// conjunction with the [WithSpanAttributes] middleware option.
type SpanAttributesFunc func(c fox.Context) []attribute.KeyValue

type config struct {
	provider   trace.TracerProvider
	propagator propagation.TextMapPropagator
	meter      metric.MeterProvider
	resolver   fox.ClientIPResolver
	carrier    func(r *http.Request) propagation.TextMapCarrier
	spanFmt    SpanNameFormatter
	attrsFn    SpanAttributesFunc
	filters    []Filter
}

func defaultConfig() *config {
	return &config{
		provider:   otel.GetTracerProvider(),
		propagator: otel.GetTextMapPropagator(),
		meter:      otel.GetMeterProvider(),
		carrier: func(r *http.Request) propagation.TextMapCarrier {
			return propagation.HeaderCarrier(r.Header)
		},
		attrsFn: func(c fox.Context) []attribute.KeyValue { return nil },
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

// WithTracerProvider specifies a tracer provider to use for tracing http request.
// If none is specified, the global tracer provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		if provider != nil {
			c.provider = provider
		}
	})
}

// WithMeterProvider specifies a meter provider to use for tracing http request.
// If none is specified, the global meter provider is used.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return optionFunc(func(c *config) {
		if provider != nil {
			c.meter = provider
		}
	})
}

// WithTextMapCarrier specify a carrier to use for extracting information from http request.
// If none is specified, [propagation.HeaderCarrier] is used.
func WithTextMapCarrier(fn func(r *http.Request) propagation.TextMapCarrier) Option {
	return optionFunc(func(c *config) {
		if fn != nil {
			c.carrier = fn
		}
	})
}

// WithSpanNameFormatter takes a function that will be called on every request
// and the returned string will become the Span Name.
func WithSpanNameFormatter(fn SpanNameFormatter) Option {
	return optionFunc(func(c *config) {
		c.spanFmt = fn
	})
}

// WithFilter adds a filter to the list of filters used by the handler. If any filter indicates to exclude a request
// then the request will not be traced. All filters must allow a request to be traced for a Span to be created.
// If no filters are provided then all requests are traced. Filters will be invoked for each processed request,
// it is advised to make them simple and fast.
func WithFilter(f ...Filter) Option {
	return optionFunc(func(c *config) {
		c.filters = append(c.filters, f...)
	})
}

// WithSpanAttributes specifies a function for generating span attributes.
// The function will be invoked for each request, and its return attributes
// will be added to the span. For example, you can use this option to add
// the http.target attribute to the span.
func WithSpanAttributes(fn SpanAttributesFunc) Option {
	return optionFunc(func(c *config) {
		if fn != nil {
			c.attrsFn = fn
		}
	})
}

// WithClientIPResolver sets a custom resolver to determine the client IP address.
// This is for advanced use case, must user should configure the resolver with Fox's router option using
// [fox.WithClientIPResolver]. Note that setting a resolver here takes priority over any resolver configured
// globally or at the route level in Fox.
func WithClientIPResolver(resolver fox.ClientIPResolver) Option {
	return optionFunc(func(c *config) {
		if resolver != nil {
			c.resolver = resolver
		}
	})
}
