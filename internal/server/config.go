package server

import "time"

type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func DefaultConfig(port int) Config {
	return Config{
		Port:            port,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    90 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}
