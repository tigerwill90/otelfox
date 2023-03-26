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

type Tracer struct {
	service    string
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	carrier    func(r *http.Request) propagation.TextMapCarrier
}

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
