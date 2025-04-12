package otelfox

import (
	"errors"
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/otelfox/internal/clientip"
	"github.com/tigerwill90/otelfox/internal/semconv"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

const (
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/tigerwill90/otelfox"
)

var (
	// DefaultClientIPResolver attempts to resolve client IP addresses in the following order:
	// 1. Leftmost non-private IP in X-Forwarded-For header
	// 2. Leftmost non-private IP in Forwarded header
	// 3. X-Real-IP header
	// 4. CF-Connecting-IP header
	// 5. True-Client-IP header
	// 6. Fastly-Client-IP header
	// 7. X-Azure-ClientIP header
	// 8. X-Appengine-Remote-Addr header
	// 9. Fly-Client-IP header
	// 10. X-Azure-SocketIP header
	// 11. RemoteAddr from the request
	//
	// The DefaultClientIPResolver uses resolvers (particularly Leftmost in X-Forwarded-For
	// and Forwarded headers) that are trivially spoofable by clients. For security-critical applications
	// where IP addresses must be trusted, consider using a Rightmost resolver or implementing
	// your own strategy tailored to your infrastructure.
	DefaultClientIPResolver = clientip.DefaultResolver
)

// Middleware returns middleware that will trace incoming requests.
// The service parameter should describe the name of the (virtual)
// server handling the request.
func Middleware(service string, opts ...Option) fox.MiddlewareFunc {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt.apply(cfg)
	}

	tracer := cfg.provider.Tracer(ScopeName, oteltrace.WithInstrumentationVersion(Version()))
	meter := cfg.meter.Meter(ScopeName, metric.WithInstrumentationVersion(Version()))

	sc := semconv.NewHTTPServer(meter)
	var hs semconv.HTTPServer

	return func(next fox.HandlerFunc) fox.HandlerFunc {
		return func(c fox.Context) {
			requestStartTime := time.Now()

			req := c.Request()

			for _, f := range cfg.filters {
				if !f(c) {
					next(c)
					return
				}
			}

			defer func() {
				// rollback to the original request
				c.SetRequest(req)
			}()

			ctx := cfg.propagator.Extract(req.Context(), cfg.carrier(req))
			requestTraceAttrOpts := semconv.RequestTraceAttrsOpts{
				HTTPClientIP: serverClientIP(c, cfg.resolver),
			}

			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(hs.RequestTraceAttrs(service, c.Request(), requestTraceAttrOpts)...),
				oteltrace.WithAttributes(hs.Route(c.Pattern())),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			var spanName string
			if cfg.spanFmt == nil {
				spanName = c.Pattern()
			} else {
				spanName = cfg.spanFmt(c)
			}
			if spanName == "" {
				spanName = scopeToString(c.Scope())
			}

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			// pass the span through the request context
			c.SetRequest(req.WithContext(ctx))

			next(c)

			status := c.Writer().Status()
			span.SetStatus(hs.Status(status))
			if status > 0 {
				span.SetAttributes(semconv.HTTPStatusCode(status))
			}

			additionalAttributes := cfg.attrsFn(c)
			sc.RecordMetrics(ctx, semconv.ServerMetricData{
				ServerName:   service,
				ResponseSize: int64(c.Writer().Size()),
				MetricAttributes: semconv.MetricAttributes{
					Req:                  c.Request(),
					StatusCode:           status,
					AdditionalAttributes: additionalAttributes,
				},
				MetricData: semconv.MetricData{
					RequestSize: c.Request().ContentLength,
					ElapsedTime: float64(time.Since(requestStartTime)) / float64(time.Millisecond),
				},
			})
		}
	}
}

func serverClientIP(c fox.Context, resolver fox.ClientIPResolver) string {
	if resolver != nil {
		ipAddr, err := resolver.ClientIP(c)
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
