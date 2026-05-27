package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/srkn0/main/internal/config"
	"github.com/srkn0/main/internal/contact"
	"github.com/srkn0/main/internal/server"
	"github.com/srkn0/main/pkg/i18n"
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

	localesSub := mustSub(localesFS, "locales")
	dataSub := mustSub(dataFS, "data")
	publicSub := mustSub(publicFS, "public")

	if err := i18n.Init(localesSub); err != nil {
		log.Fatalf("i18n init: %v", err)
	}

	contactSvc := contact.NewService(buildMailer(cfg), contact.Config{
		From: cfg.SMTP.From,
		To:   cfg.SMTP.To,
	})

	if err := server.Run(dataSub, publicSub, server.DefaultConfig(cfg.Port), contactSvc); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func buildMailer(cfg config.Config) contact.Mailer {
	if cfg.SMTP.Host == "" {
		log.Println("mail: SMTP_HOST not set, using stdout mailer (dev mode)")
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
