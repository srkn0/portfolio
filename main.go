package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"os"

	"github.com/srkn0/main/internal/config"
	"github.com/srkn0/main/internal/contact"
	"github.com/srkn0/main/internal/o11y"
	"github.com/srkn0/main/internal/server"
	"github.com/srkn0/main/pkg/i18n"
)

// ldflags-injectable build info. CI sets these at build time:
//
//	go build -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT"
var (
	version = "dev"
	commit  = "unknown"
)

//go:embed all:locales
var localesFS embed.FS

//go:embed all:data
var dataFS embed.FS

//go:embed all:public
var publicFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	logger, logShutdown, err := o11y.NewLogger(ctx, o11y.LoggerConfig{
		ServiceName:  cfg.Obs.ServiceName,
		OTLPEndpoint: cfg.Obs.OTLPEndpoint,
		Level:        o11y.ParseLevel(cfg.Obs.LogLevel),
		Writer:       os.Stderr,
	})
	if err != nil {
		log.Fatalf("logger init: %v", err)
	}
	slog.SetDefault(logger)
	defer func() { _ = logShutdown(context.Background()) }()

	traceShutdown, err := o11y.NewTracer(ctx, o11y.TraceConfig{
		ServiceName:  cfg.Obs.ServiceName,
		OTLPEndpoint: cfg.Obs.OTLPEndpoint,
	})
	if err != nil {
		logger.Error("tracer init", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = traceShutdown(context.Background()) }()

	recorder := o11y.NewRecorder(o11y.DefaultRegistry())
	recorder.SetBuildInfo(version, commit)

	localesSub := mustSub(localesFS, "locales")
	dataSub := mustSub(dataFS, "data")
	publicSub := mustSub(publicFS, "public")

	if err := i18n.Init(localesSub); err != nil {
		logger.Error("i18n init", slog.String("err", err.Error()))
		os.Exit(1)
	}

	contactSvc := contact.NewService(buildMailer(cfg, logger), contact.Config{
		From: cfg.SMTP.From,
		To:   cfg.SMTP.To,
	})

	err = server.Run(dataSub, publicSub, server.DefaultConfig(cfg.Port), server.Deps{
		Logger:     logger,
		Recorder:   recorder,
		ContactSvc: contactSvc,
		Version:    version,
		Commit:     commit,
	})
	if err != nil {
		logger.Error("server", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func buildMailer(cfg config.Config, logger *slog.Logger) contact.Mailer {
	if cfg.SMTP.Host == "" {
		logger.Info("mail.fallback",
			slog.String("event", "mail_fallback"),
			slog.String("reason", "SMTP_HOST not set, using stdout mailer (dev mode)"),
		)
		return contact.StdoutMailer{}
	}
	return contact.NewSMTPMailer(cfg.SMTP)
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		log.Fatalf("sub fs %q: %v", dir, err)
	}
	return sub
}
