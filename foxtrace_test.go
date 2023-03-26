package otelfox

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigerwill90/fox"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"testing"

	b3prop "go.opentelemetry.io/contrib/propagators/b3"
)

func TestGetSpanNotInstrumented(t *testing.T) {
	router := fox.New()
	require.NoError(t, router.Handler(http.MethodGet, "/ping", fox.HandlerFunc(func(w http.ResponseWriter, r *http.Request, parms fox.Params) {
		span := trace.SpanFromContext(r.Context())
		ok := !span.SpanContext().IsValid()
		assert.True(t, ok)
		_, _ = fmt.Fprint(w, "ok")
	})))

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	response := w.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestPropagationWithGlobalPropagators(t *testing.T) {
	provider := trace.NewNoopTracerProvider()
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
	err := router.Handler(http.MethodGet, "/user/:id", mw.Middleware(fox.HandlerFunc(func(w http.ResponseWriter, r *http.Request, params fox.Params) {
		span := trace.SpanFromContext(r.Context())
		assert.Equal(t, sc.TraceID(), span.SpanContext().TraceID())
		assert.Equal(t, sc.SpanID(), span.SpanContext().SpanID())
	})))

	require.NoError(t, err)
	router.ServeHTTP(w, r)
}

func TestPropagationWithCustomPropagators(t *testing.T) {
	provider := trace.NewNoopTracerProvider()
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
	err := router.Handler(http.MethodGet, "/user/:id", mw.Middleware(fox.HandlerFunc(func(w http.ResponseWriter, r *http.Request, params fox.Params) {
		span := trace.SpanFromContext(r.Context())
		assert.Equal(t, sc.TraceID(), span.SpanContext().TraceID())
		assert.Equal(t, sc.SpanID(), span.SpanContext().SpanID())
	})))

	require.NoError(t, err)
	router.ServeHTTP(w, r)
}
