package o11y_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/srkn0/main/internal/o11y"
)

func newCapturingLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func TestLogStartup_emitsStructuredRecord(t *testing.T) {
	var buf bytes.Buffer
	logger := newCapturingLogger(&buf)

	o11y.LogStartup(logger, o11y.StartupInfo{
		Version:      "v0.1.0",
		Commit:       "abc1234",
		Addr:         ":8080",
		PostCount:    12,
		ProjectCount: 1,
	})

	var record map[string]any
	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("LogStartup wrote nothing")
	}
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, line)
	}

	for _, want := range []string{"v0.1.0", "abc1234", ":8080"} {
		if !valueAnywhere(record, want) {
			t.Errorf("startup record missing %q\n%s", want, line)
		}
	}
}

func TestLogStartup_hasStableEventName(t *testing.T) {
	// Loki queries should be able to filter on a stable event name.
	// We don't care exactly what it's called, just that there IS one
	// and that it's "startup-y" (so it's discoverable in dashboards).
	var buf bytes.Buffer
	o11y.LogStartup(newCapturingLogger(&buf), o11y.StartupInfo{Addr: ":8080"})

	out := buf.String()
	if !strings.Contains(out, "startup") && !strings.Contains(out, "boot") {
		t.Errorf("expected the record to carry a startup-ish event name, got:\n%s", out)
	}
}

func TestLogShutdown_includesDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := newCapturingLogger(&buf)

	o11y.LogShutdown(logger, 142*time.Millisecond)

	out := buf.String()
	if out == "" {
		t.Fatal("LogShutdown wrote nothing")
	}
	// Either "142ms" (slog.Duration default), "0.142s", or a "duration"/"drained" key.
	if !strings.Contains(out, "142") {
		t.Errorf("expected the drained duration in the output, got:\n%s", out)
	}
}

// valueAnywhere walks a JSON map and looks for `want` as a string anywhere.
// Used so the test doesn't pin a specific attribute key.
func valueAnywhere(v any, want string) bool {
	switch x := v.(type) {
	case string:
		return strings.Contains(x, want)
	case float64:
		// Numbers don't match strings we look for.
		return false
	case map[string]any:
		for _, sub := range x {
			if valueAnywhere(sub, want) {
				return true
			}
		}
	case []any:
		for _, sub := range x {
			if valueAnywhere(sub, want) {
				return true
			}
		}
	}
	return false
}
