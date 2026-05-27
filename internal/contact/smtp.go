package contact

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/srkn0/main/internal/config"
)

type SMTPMailer struct {
	cfg config.SMTPConfig
}

func NewSMTPMailer(cfg config.SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(_ context.Context, mail Mail) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	return smtp.SendMail(addr, auth, mail.From, mail.To, buildMessage(mail))
}

func buildMessage(m Mail) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", m.From)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(m.To, ", "))
	if m.ReplyTo != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", m.ReplyTo)
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", m.Subject)
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=utf-8\r\n")
	fmt.Fprintf(&b, "\r\n")
	b.WriteString(m.Body)
	return []byte(b.String())
}
