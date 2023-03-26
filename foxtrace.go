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

// Middleware returns a Fox middleware function that traces HTTP requests.
// The span name for each request is retrieved from fox.Params using params.Get(fox.RouteKey).
// If the matched route is not found in params, the span name is set to "HTTP {method} route not found".
func (t *Tracer) Middleware(h fox.Handler) fox.Handler {
	return fox.HandlerFunc(func(w http.ResponseWriter, r *http.Request, params fox.Params) {
		ctx := t.propagator.Extract(r.Context(), t.carrier(r))
		spanName := params.Get(fox.RouteKey)
		opts := []trace.SpanStartOption{
			trace.WithAttributes(httpconv.ServerRequest(t.service, r)...),
			trace.WithSpanKind(trace.SpanKindServer),
		}
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", r.Method)
		} else {
			opts = append(opts, trace.WithAttributes(semconv.HTTPRoute(spanName)))
		}

		ctx, span := t.tracer.Start(ctx, spanName, opts...)
		defer span.End()

		recorder := newResponseStatusRecorder(w)
		defer recorder.free()

		h.ServeHTTP(recorder.w, r.WithContext(ctx), params)
		status := recorder.status

		span.SetStatus(httpconv.ServerStatus(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	})
}
