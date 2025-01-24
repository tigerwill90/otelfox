package otelfox

import (
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/fox/clientip"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

// Filter is a function that determines whether a given HTTP request should be traced.
// It returns true to indicate the request should not be traced.
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
		carrier: func(r *http.Request) propagation.TextMapCarrier {
			return propagation.HeaderCarrier(r.Header)
		},
		resolver: clientip.NewChain(
			must(clientip.NewLeftmostNonPrivate(clientip.XForwardedForKey, 15)),
			must(clientip.NewLeftmostNonPrivate(clientip.ForwardedKey, 15)),
			must(clientip.NewSingleIPHeader(fox.HeaderCFConnectionIP)),
			must(clientip.NewSingleIPHeader(fox.HeaderTrueClientIP)),
			must(clientip.NewSingleIPHeader(fox.HeaderFastClientIP)),
			must(clientip.NewSingleIPHeader(fox.HeaderXAzureClientIP)),
			must(clientip.NewSingleIPHeader(fox.HeaderXAzureSocketIP)),
			must(clientip.NewSingleIPHeader(fox.HeaderXAppengineRemoteAddr)),
			must(clientip.NewSingleIPHeader(fox.HeaderFlyClientIP)),
			must(clientip.NewSingleIPHeader(fox.HeaderXRealIP)),
			clientip.NewRemoteAddr(),
		),
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

// WithFilter appends the provided filters to the middleware's filter list.
// A filter returning true will exclude the request from being traced. If no filters
// are provided, all requests will be traced. Keep in mind that filters are invoked for each request,
// so they should be simple and efficient.
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
		c.attrsFn = fn
	})
}

// WithClientIPResolver sets a custom resolver to determine the client IP address.
// This is for advanced use case, must user should configure the resolver with Fox's router option using
// [fox.WithClientIPResolver].
func WithClientIPResolver(resolver fox.ClientIPResolver) Option {
	return optionFunc(func(c *config) {
		if resolver != nil {
			c.resolver = resolver
		}
	})
}

func must(resolver fox.ClientIPResolver, err error) fox.ClientIPResolver {
	if err != nil {
		panic(err)
	}
	return resolver
}
