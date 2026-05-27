package o11y

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

// TraceConfig configures the OTel TracerProvider. When OTLPEndpoint is empty,
// NewTracer returns a no-op shutdown and the global tracer provider stays a no-op,
// so adding TraceMiddleware everywhere doesn't pull in network calls in dev.
type TraceConfig struct {
	ServiceName  string
	OTLPEndpoint string
}

// NewTracer wires up the global OTel TracerProvider with an OTLP/HTTP exporter.
// Returns a shutdown function that flushes pending spans on app exit.
func NewTracer(ctx context.Context, cfg TraceConfig) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }
	if cfg.OTLPEndpoint == "" {
		return noop, nil
	}

	host, insecure, err := parseOTLPEndpoint(cfg.OTLPEndpoint)
	if err != nil {
		return nil, err
	}
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(host)}
	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(opts...))
	if err != nil {
		return nil, fmt.Errorf("otlp trace exporter: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(cfg.ServiceName),
	))
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}

// TraceMiddleware wraps the router with otelhttp so every request gets a span,
// and after the chi route matches it renames the span to "METHOD /pattern" plus
// adds the http.route attribute. Use this OUTSIDE the chi router so 404s are also
// traced.
func TraceMiddleware(next http.Handler) http.Handler {
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if pattern := RouteFromContext(r); pattern != "" {
			span := trace.SpanFromContext(r.Context())
			span.SetName(r.Method + " " + pattern)
			span.SetAttributes(attribute.String("http.route", pattern))
		}
	})
	return otelhttp.NewHandler(wrapped, "http.request")
}
