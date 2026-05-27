package contact

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidForm = errors.New("invalid contact form")

type Form struct {
	Name    string
	Email   string
	Subject string
	Message string
}

type Mail struct {
	From    string
	To      []string
	ReplyTo string
	Subject string
	Body    string
}

type Mailer interface {
	Send(ctx context.Context, m Mail) error
}

type Config struct {
	From string
	To   string
}

type Service struct {
	mailer Mailer
	cfg    Config
}

func NewService(mailer Mailer, cfg Config) *Service {
	return &Service{mailer: mailer, cfg: cfg}
}

func (s *Service) Send(ctx context.Context, form Form) error {
	if err := validateForm(form); err != nil {
		return err
	}
	mail := Mail{
		From:    s.cfg.From,
		To:      []string{s.cfg.To},
		ReplyTo: form.Email,
		Subject: form.Subject,
		Body: fmt.Sprintf(
			"From: %s <%s>\n\n%s",
			form.Name, form.Email, form.Message,
		),
	}
	return s.mailer.Send(ctx, mail)
}

func validateForm(f Form) error {
	if f.Name == "" || f.Email == "" || f.Subject == "" || f.Message == "" {
		return fmt.Errorf("%w: all fields required", ErrInvalidForm)
	}
	if !strings.Contains(f.Email, "@") {
		return fmt.Errorf("%w: invalid email", ErrInvalidForm)
	}
	return nil
}
