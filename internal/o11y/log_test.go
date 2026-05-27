package o11y_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/srkn0/main/internal/o11y"
)

func TestNewLogger_writesJSON(t *testing.T) {
	var buf bytes.Buffer
	logger, shutdown, err := o11y.NewLogger(context.Background(), o11y.LoggerConfig{
		ServiceName: "portfolio",
		Level:       slog.LevelInfo,
		Writer:      &buf,
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	logger.Info("hello", "k", "v")

	line := strings.TrimSpace(buf.String())
	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("output is not JSON: %v\nline: %s", err, line)
	}
	if record["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", record["msg"])
	}
	if record["k"] != "v" {
		t.Errorf("k = %v, want v", record["k"])
	}
}

func TestNewLogger_respectsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger, _, err := o11y.NewLogger(context.Background(), o11y.LoggerConfig{
		ServiceName: "portfolio",
		Level:       slog.LevelWarn,
		Writer:      &buf,
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	logger.Info("should not appear")
	logger.Warn("should appear")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Error("Info logged despite Warn level")
	}
	if !strings.Contains(out, "should appear") {
		t.Error("Warn record missing")
	}
}

func TestNewLogger_includesServiceName(t *testing.T) {
	// The service name must end up on every record so it survives the
	// OTLP export and lets you filter by service in Grafana/Loki.
	var buf bytes.Buffer
	logger, _, err := o11y.NewLogger(context.Background(), o11y.LoggerConfig{
		ServiceName: "portfolio",
		Level:       slog.LevelInfo,
		Writer:      &buf,
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	logger.Info("boot")

	var record map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &record); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if !hasValueDeep(record, "portfolio") {
		t.Errorf("expected service.name=portfolio somewhere in record, got %v", record)
	}
}

func TestNewLogger_shutdownDoesNotPanic(t *testing.T) {
	logger, shutdown, err := o11y.NewLogger(context.Background(), o11y.LoggerConfig{
		ServiceName: "portfolio",
		Level:       slog.LevelInfo,
		Writer:      bytes.NewBuffer(nil),
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	logger.Info("flush me")

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown returned error: %v", err)
	}
}

// hasValueDeep walks a decoded JSON record and looks for the given string value
// anywhere in the tree. Tolerates whichever attribute key the implementation
// picks (service, service.name, etc).
func hasValueDeep(v any, want string) bool {
	switch x := v.(type) {
	case string:
		return x == want
	case map[string]any:
		for _, sub := range x {
			if hasValueDeep(sub, want) {
				return true
			}
		}
	case []any:
		for _, sub := range x {
			if hasValueDeep(sub, want) {
				return true
			}
		}
	}
	return false
}
