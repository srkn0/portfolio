package o11y

import (
	"log/slog"
	"time"
)

// StartupInfo carries the fields stamped on the boot log line.
// The same fields are useful as labels on the build_info Prometheus gauge.
type StartupInfo struct {
	Version      string
	Commit       string
	Addr         string
	PostCount    int
	ProjectCount int
}

// LogStartup writes one structured Info record summarising the boot. Event
// name is "startup" so Loki queries can filter by event="startup".
func LogStartup(logger *slog.Logger, info StartupInfo) {
	logger.Info("startup",
		slog.String("event", "startup"),
		slog.String("version", info.Version),
		slog.String("commit", info.Commit),
		slog.String("addr", info.Addr),
		slog.Int("post_count", info.PostCount),
		slog.Int("project_count", info.ProjectCount),
	)
}

// LogShutdown writes one structured Info record on graceful shutdown,
// including how long the drain took.
func LogShutdown(logger *slog.Logger, drained time.Duration) {
	logger.Info("shutdown",
		slog.String("event", "shutdown"),
		slog.Duration("drained", drained),
	)
}
