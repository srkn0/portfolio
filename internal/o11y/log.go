package o11y

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// parseOTLPEndpoint splits the OTEL_EXPORTER_OTLP_ENDPOINT env value into
// the host:port portion and a flag for whether the connection is insecure
// (HTTP scheme). The SDK appends the signal-specific path (/v1/logs,
// /v1/traces) automatically when given just the host, so a base URL like
// http://otel-collector:4318 routes correctly to all signals.
func parseOTLPEndpoint(raw string) (host string, insecure bool, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false, fmt.Errorf("parse OTLP endpoint %q: %w", raw, err)
	}
	return u.Host, u.Scheme != "https", nil
}

// LoggerConfig carries everything NewLogger needs.
//
//	OTLPEndpoint — when set, log records also export via OTLP/HTTP to that endpoint
//	               (e.g. http://otel-collector:4318). When empty, only local JSON.
//	ServiceName — added as service.name on every record, becomes a Loki label.
//	Level       — minimum slog level emitted.
//	Writer      — destination for local JSON; defaults to os.Stderr if nil.
type LoggerConfig struct {
	ServiceName  string
	OTLPEndpoint string
	Level        slog.Level
	Writer       io.Writer
}

// NewLogger builds an *slog.Logger that emits structured JSON locally and,
// if OTLPEndpoint is set, also exports records via OTLP.
//
// The returned shutdown function flushes pending records; call it on app exit.
func NewLogger(ctx context.Context, cfg LoggerConfig) (*slog.Logger, func(context.Context) error, error) {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}

	jsonHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: cfg.Level})

	handlers := []slog.Handler{jsonHandler}
	shutdown := func(context.Context) error { return nil }

	if cfg.OTLPEndpoint != "" {
		host, insecure, err := parseOTLPEndpoint(cfg.OTLPEndpoint)
		if err != nil {
			return nil, nil, err
		}
		opts := []otlploghttp.Option{otlploghttp.WithEndpoint(host)}
		if insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		exporter, err := otlploghttp.New(ctx, opts...)
		if err != nil {
			return nil, nil, fmt.Errorf("otlp log exporter: %w", err)
		}

		res, err := resource.New(ctx, resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		))
		if err != nil {
			return nil, nil, fmt.Errorf("otel resource: %w", err)
		}

		provider := sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
			sdklog.WithResource(res),
		)

		otelHandler := otelslog.NewHandler(cfg.ServiceName, otelslog.WithLoggerProvider(provider))
		handlers = append(handlers, otelHandler)
		shutdown = provider.Shutdown
	}

	logger := slog.New(&fanoutHandler{handlers: handlers}).With(
		slog.String("service.name", cfg.ServiceName),
	)
	return logger, shutdown, nil
}

// fanoutHandler dispatches each record to every contained handler.
// Used when both local JSON and OTLP export are active.
type fanoutHandler struct {
	handlers []slog.Handler
}

func (f *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range f.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (f *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	subs := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		subs[i] = h.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: subs}
}

func (f *fanoutHandler) WithGroup(name string) slog.Handler {
	subs := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		subs[i] = h.WithGroup(name)
	}
	return &fanoutHandler{handlers: subs}
}

// ParseLevel converts a string like "debug" / "info" / "warn" / "error" to
// the matching slog.Level. Unknown or empty input falls back to Info so the
// app never refuses to start over a typo in env config.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
