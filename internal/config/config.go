package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var ErrMissingEnv = errors.New("missing required environment variable")

type Config struct {
	Port int
	SMTP SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       string
}

type envSpec struct {
	key      string
	target   *string
	required bool
	fallback string
}

func loadSpecs(specs []envSpec) error {
	for _, s := range specs {
		v, ok := os.LookupEnv(s.key)
		if !ok || v == "" {
			if s.required {
				return fmt.Errorf("%w: %s", ErrMissingEnv, s.key)
			}
			v = s.fallback
		}
		*s.target = v
	}
	return nil
}

func Load() (Config, error) {
	var cfg Config
	var portStr, smtpPortStr string

	if err := loadSpecs([]envSpec{
		{key: "PORT", target: &portStr, fallback: "8080"},
	}); err != nil {
		return Config{}, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("PORT: %w", err)
	}
	cfg.Port = port

	host, ok := os.LookupEnv("SMTP_HOST")
	if !ok || host == "" {
		return cfg, nil
	}
	cfg.SMTP.Host = host

	if err := loadSpecs([]envSpec{
		{key: "SMTP_PORT", target: &smtpPortStr, fallback: "587"},
		{key: "SMTP_USERNAME", target: &cfg.SMTP.Username, required: true},
		{key: "SMTP_PASSWORD", target: &cfg.SMTP.Password, required: true},
		{key: "SMTP_FROM", target: &cfg.SMTP.From, required: true},
		{key: "SMTP_TO", target: &cfg.SMTP.To, required: true},
	}); err != nil {
		return Config{}, err
	}
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return Config{}, fmt.Errorf("SMTP_PORT: %w", err)
	}
	cfg.SMTP.Port = smtpPort

	return cfg, nil
}
