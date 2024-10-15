package otelfox

import (
	"fmt"
	"github.com/tigerwill90/fox"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "github.com/tigerwill90/otelfox"
)

// Tracer is a Fox middleware that traces HTTP requests using OpenTelemetry.
type Tracer struct {
	service string
	tracer  trace.Tracer
	cfg     *config
}

// New creates a new Tracer middleware for the given service.
// Options can be provided to configure the tracer.
func New(service string, opts ...Option) *Tracer {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt.apply(cfg)
	}

	tracer := cfg.provider.Tracer(tracerName, trace.WithInstrumentationVersion(SemVersion()))
	return &Tracer{
		service: service,
		tracer:  tracer,
		cfg:     cfg,
	}
}

// Middleware is a convenience function that creates a new Tracer middleware instance
// for the specified service and returns the Trace middleware function.
// Options can be provided to configure the tracer.
func Middleware(service string, opts ...Option) fox.MiddlewareFunc {
	tracer := New(service, opts...)
	return tracer.Trace
}

// Trace is a middleware function that wraps the provided HandlerFunc with tracing capabilities.
// It captures and records HTTP request information using OpenTelemetry.
func (t *Tracer) Trace(next fox.HandlerFunc) fox.HandlerFunc {
	return func(c fox.Context) {

		req := c.Request()

		for _, f := range t.cfg.filters {
			if f(c) {
				next(c)
				return
			}
		}

		savedCtx := req.Context()
		defer func() {
			// rollback to the original context
			c.SetRequest(req.WithContext(savedCtx))
		}()

		ctx := t.cfg.propagator.Extract(savedCtx, t.cfg.carrier(req))

		attributes := httpconv.ServerRequest(t.service, req)
		if t.cfg.attrsFn != nil {
			attributes = append(attributes, t.cfg.attrsFn(c)...)
		}

		opts := []trace.SpanStartOption{
			trace.WithAttributes(attributes...),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		var spanName string
		if t.cfg.spanFmt == nil {
			spanName = c.Path()
		} else {
			spanName = t.cfg.spanFmt(c)
		}

		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", req.Method)
		} else {
			opts = append(opts, trace.WithAttributes(semconv.HTTPRoute(spanName)))
		}

		ctx, span := t.tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.SetRequest(req.WithContext(ctx))

		next(c)

		status := c.Writer().Status()
		span.SetStatus(httpconv.ServerStatus(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	}
}
