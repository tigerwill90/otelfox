package otelfox

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigerwill90/fox"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"net/http"
	"net/http/httptest"
	"testing"

	b3prop "go.opentelemetry.io/contrib/propagators/b3"
)

func TestGetSpanNotInstrumented(t *testing.T) {
	router := fox.New()
	_, err := router.Handle(http.MethodGet, "/ping", func(c fox.Context) {
		span := trace.SpanFromContext(c.Request().Context())
		ok := !span.SpanContext().IsValid()
		assert.True(t, ok)
		_ = c.String(http.StatusOK, "ok")
	})
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	response := w.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestPropagationWithGlobalPropagators(t *testing.T) {
	provider := noop.NewTracerProvider()
	otel.SetTextMapPropagator(b3prop.New())

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
	ctx, _ = provider.Tracer(tracerName).Start(ctx, "test")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	router := fox.New()
	mw := New("foobar", WithTracerProvider(provider))
	_, err := router.Handle(http.MethodGet, "/user/{id}", mw.Trace(func(c fox.Context) {
		span := trace.SpanFromContext(c.Request().Context())
		assert.Equal(t, sc.TraceID(), span.SpanContext().TraceID())
		assert.Equal(t, sc.SpanID(), span.SpanContext().SpanID())
	}))

	require.NoError(t, err)
	router.ServeHTTP(w, r)
}

func TestPropagationWithCustomPropagators(t *testing.T) {
	provider := noop.NewTracerProvider()
	b3 := b3prop.New()

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
	ctx, _ = provider.Tracer(tracerName).Start(ctx, "test")
	b3.Inject(ctx, propagation.HeaderCarrier(r.Header))

	router := fox.New()
	mw := New("foobar", WithTracerProvider(provider), WithPropagators(b3))
	_, err := router.Handle(http.MethodGet, "/user/{id}", mw.Trace(func(c fox.Context) {
		span := trace.SpanFromContext(c.Request().Context())
		assert.Equal(t, sc.TraceID(), span.SpanContext().TraceID())
		assert.Equal(t, sc.SpanID(), span.SpanContext().SpanID())
	}))

	require.NoError(t, err)
	router.ServeHTTP(w, r)
}

func TestWithSpanAttributes(t *testing.T) {
	provider := noop.NewTracerProvider()
	otel.SetTextMapPropagator(b3prop.New())

	r := httptest.NewRequest("GET", "/user/123?foo=bar", nil)
	w := httptest.NewRecorder()

	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
	ctx, _ = provider.Tracer(tracerName).Start(ctx, "test")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	router := fox.New()
	mw := New("foobar", WithTracerProvider(provider), WithSpanAttributes(func(c fox.Context) []attribute.KeyValue {
		attrs := make([]attribute.KeyValue, 1, 2)
		attrs[0] = attribute.String("http.target", r.URL.String())
		for annotation := range c.Route().Annotations() {
			attrs = append(attrs, attribute.KeyValue{
				Key:   attribute.Key(annotation.Key),
				Value: attribute.StringValue(annotation.Value.(string)),
			})
		}
		return attrs
	}))
	_, err := router.Handle(http.MethodGet, "/user/{id}", mw.Trace(func(c fox.Context) {
		span := trace.SpanFromContext(c.Request().Context())
		assert.Equal(t, sc.TraceID(), span.SpanContext().TraceID())
		assert.Equal(t, sc.SpanID(), span.SpanContext().SpanID())
	}), fox.WithAnnotations(fox.Annotation{Key: "foo", Value: "bar"}))

	require.NoError(t, err)
	router.ServeHTTP(w, r)
}
