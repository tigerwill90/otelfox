package otelfox

import (
	"fmt"
	"github.com/tigerwill90/fox"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/semconv/v1.18.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

const (
	tracerName = "github.com/tigerwill90/otelfox"
)

// Tracer is a Fox middleware that traces HTTP requests using OpenTelemetry.
type Tracer struct {
	service    string
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	carrier    func(r *http.Request) propagation.TextMapCarrier
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
		service:    service,
		tracer:     tracer,
		propagator: cfg.propagator,
		carrier:    cfg.carrier,
	}
}

func (t *Tracer) Trace(next fox.HandlerFunc) fox.HandlerFunc {
	return func(c fox.Context) {

		req := c.Request()
		ctx := t.propagator.Extract(req.Context(), t.carrier(req))

		spanName := c.Path()
		opts := []trace.SpanStartOption{
			trace.WithAttributes(httpconv.ServerRequest(t.service, c.Request())...),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request().Method)
		} else {
			opts = append(opts, trace.WithAttributes(semconv.HTTPRoute(spanName)))
		}

		ctx, span := t.tracer.Start(ctx, spanName, opts...)
		defer span.End()

		c.SetRequest(c.Request().WithContext(ctx))

		next(c)

		status := c.Writer().Status()
		span.SetStatus(httpconv.ServerStatus(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	}
}
