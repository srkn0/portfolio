package o11y

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Stable event names. Use these as the `event` attribute when calling LogEvent
// so Loki/Grafana queries pivot on a known set: {event="contact_submit"}.
const (
	EventContactSubmit     = "contact_submit"
	EventContactValidation = "contact_validation_failed"
	EventContactSendFailed = "contact_send_failed"
	EventLangSwitch        = "lang_switch"
)

// LogEvent records a domain event into three places at once:
//   - slog (JSON → stdout + Loki via OTLP)
//   - the active OTel span as a span event (visible in Tempo under the parent)
//   - includes trace_id on the slog record so Loki ↔ Tempo correlation works
//
// Pass extra context via attrs.
func LogEvent(ctx context.Context, logger *slog.Logger, event string, attrs ...slog.Attr) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		kvs := make([]attribute.KeyValue, 0, len(attrs))
		for _, a := range attrs {
			kvs = append(kvs, slogAttrToKV(a))
		}
		span.AddEvent(event, trace.WithAttributes(kvs...))
	}

	enriched := append(make([]slog.Attr, 0, len(attrs)+2),
		slog.String("event", event),
	)
	if tid := traceIDFrom(ctx); tid != "" {
		enriched = append(enriched, slog.String("trace_id", tid))
	}
	enriched = append(enriched, attrs...)

	logger.LogAttrs(ctx, slog.LevelInfo, event, enriched...)
}

// traceIDFrom returns the active span's trace ID, or "" when there isn't one.
// Used to stamp log records with the trace ID for Loki ↔ Tempo correlation.
func traceIDFrom(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

func slogAttrToKV(a slog.Attr) attribute.KeyValue {
	switch a.Value.Kind() {
	case slog.KindString:
		return attribute.String(a.Key, a.Value.String())
	case slog.KindInt64:
		return attribute.Int64(a.Key, a.Value.Int64())
	case slog.KindBool:
		return attribute.Bool(a.Key, a.Value.Bool())
	case slog.KindFloat64:
		return attribute.Float64(a.Key, a.Value.Float64())
	case slog.KindDuration:
		return attribute.Int64(a.Key, a.Value.Duration().Milliseconds())
	default:
		return attribute.String(a.Key, a.Value.String())
	}
}
