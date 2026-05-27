package config_test

import (
	"errors"
	"testing"

	"github.com/srkn0/main/internal/config"
)

func setSMTPEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SMTP_HOST", "smtp.gmail.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "me@example.com")
	t.Setenv("SMTP_PASSWORD", "app-password")
	t.Setenv("SMTP_FROM", "me@example.com")
	t.Setenv("SMTP_TO", "inbox@example.com")
}

func TestLoad_portDefault(t *testing.T) {
	t.Setenv("PORT", "")
	setSMTPEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
}

func TestLoad_portFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	setSMTPEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
}

func TestLoad_portInvalidReturnsError(t *testing.T) {
	t.Setenv("PORT", "not-a-number")
	setSMTPEnv(t)

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}

func TestLoad_smtpAllFieldsPresent(t *testing.T) {
	setSMTPEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := config.SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "me@example.com",
		Password: "app-password",
		From:     "me@example.com",
		To:       "inbox@example.com",
	}
	if cfg.SMTP != want {
		t.Errorf("SMTP = %+v\nwant     %+v", cfg.SMTP, want)
	}
}

func TestLoad_smtpPortDefault(t *testing.T) {
	setSMTPEnv(t)
	t.Setenv("SMTP_PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SMTP.Port != 587 {
		t.Errorf("Port = %d, want 587", cfg.SMTP.Port)
	}
}

func TestLoad_smtpPortInvalidReturnsError(t *testing.T) {
	setSMTPEnv(t)
	t.Setenv("SMTP_PORT", "not-a-number")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid SMTP_PORT")
	}
}

func TestLoad_missingRequiredKey(t *testing.T) {
	// Once SMTP_HOST is given, the rest of the SMTP_* keys are required.
	// SMTP_USERNAME left empty here.
	t.Setenv("SMTP_HOST", "smtp.gmail.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "x")
	t.Setenv("SMTP_FROM", "x")
	t.Setenv("SMTP_TO", "x")

	_, err := config.Load()
	if !errors.Is(err, config.ErrMissingEnv) {
		t.Errorf("err = %v, want errors.Is(err, ErrMissingEnv)", err)
	}
}

func TestLoad_smtpDisabledWhenHostEmpty(t *testing.T) {
	// If SMTP_HOST is empty, Load should not complain about the other
	// SMTP_* keys — the app should still start without a mailer.
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_TO", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load with empty SMTP must succeed: %v", err)
	}
	if cfg.SMTP.Host != "" {
		t.Errorf("Host = %q, want empty", cfg.SMTP.Host)
	}
}
