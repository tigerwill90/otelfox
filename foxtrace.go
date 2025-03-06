package otelfox

import (
	"errors"
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/otelfox/internal/clientip"
	"github.com/tigerwill90/otelfox/internal/semconvutil"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "github.com/tigerwill90/otelfox"
)

type middleware struct {
	tracer  trace.Tracer
	cfg     *config
	service string
}

// Middleware returns middleware that will trace incoming requests.
// The service parameter should describe the name of the (virtual)
// server handling the request.
func Middleware(service string, opts ...Option) fox.MiddlewareFunc {
	tracer := createTracer(service, opts...)
	return tracer.trace
}

func createTracer(service string, opts ...Option) middleware {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt.apply(cfg)
	}

	tracer := cfg.provider.Tracer(tracerName, trace.WithInstrumentationVersion(SemVersion()))
	return middleware{
		service: service,
		tracer:  tracer,
		cfg:     cfg,
	}
}

func (t middleware) trace(next fox.HandlerFunc) fox.HandlerFunc {
	return func(c fox.Context) {

		req := c.Request()

		for _, f := range t.cfg.filters {
			if f(c) {
				next(c)
				return
			}
		}

		defer func() {
			// rollback to the original request
			c.SetRequest(req)
		}()

		ctx := t.cfg.propagator.Extract(req.Context(), t.cfg.carrier(req))

		attributes := semconvutil.HTTPServerRequest(t.service, req)
		if t.cfg.attrsFn != nil {
			attributes = append(attributes, t.cfg.attrsFn(c)...)
		}

		opts := make([]trace.SpanStartOption, 0, 5)
		opts = append(
			opts,
			trace.WithAttributes(attributes...),
			trace.WithAttributes(semconv.HTTPRoute(c.Pattern())),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		clientIp := t.serverClientIP(c)
		if clientIp != "" {
			opts = append(opts, trace.WithAttributes(semconv.HTTPClientIP(clientIp)))
		}

		var spanName string
		if t.cfg.spanFmt == nil {
			spanName = c.Pattern()
		} else {
			spanName = t.cfg.spanFmt(c)
		}

		if spanName == "" {
			spanName = scopeToString(c.Scope())
		}

		opts = append(opts, trace.WithAttributes(semconv.HTTPRoute(spanName)))
		ctx, span := t.tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.SetRequest(req.WithContext(ctx))

		next(c)

		status := c.Writer().Status()
		span.SetStatus(semconvutil.HTTPServerStatus(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	}
}

func (t middleware) serverClientIP(c fox.Context) string {
	if t.cfg.resolver != nil {
		ipAddr, err := t.cfg.resolver.ClientIP(c)
		if err != nil {
			return ""
		}
		return ipAddr.String()
	}

	ipAddr, err := c.ClientIP()
	if err == nil {
		return ipAddr.String()
	}
	if errors.Is(err, fox.ErrNoClientIPResolver) {
		ipAddr, err = clientip.DefaultResolver.ClientIP(c)
		if err == nil {
			return ipAddr.String()
		}
	}
	return ""
}

func scopeToString(scope fox.HandlerScope) string {
	var strScope string
	switch scope {
	case fox.OptionsHandler:
		strScope = "OptionsHandler"
	case fox.NoMethodHandler:
		strScope = "NoMethodHandler"
	case fox.RedirectHandler:
		strScope = "RedirectHandler"
	case fox.NoRouteHandler:
		strScope = "NoRouteHandler"
	default:
		strScope = "UnknownHandler"
	}
	return strScope
}
