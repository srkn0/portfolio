package contact_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/srkn0/main/internal/contact"
)

type spyMailer struct {
	called bool
	mail   contact.Mail
	err    error
}

func (s *spyMailer) Send(_ context.Context, m contact.Mail) error {
	s.called = true
	s.mail = m
	return s.err
}

func newService(mailer contact.Mailer) *contact.Service {
	return contact.NewService(mailer, contact.Config{
		From: "site@example.com",
		To:   "inbox@example.com",
	})
}

func validForm() contact.Form {
	return contact.Form{
		Name:    "Max Mustermann",
		Email:   "max@example.com",
		Subject: "Frage zum Projekt",
		Message: "Hi, ich hätte da eine Frage zu deinem Setup.",
	}
}

func TestService_Send_callsMailerOnce(t *testing.T) {
	mailer := &spyMailer{}
	svc := newService(mailer)

	if err := svc.Send(context.Background(), validForm()); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !mailer.called {
		t.Fatal("expected Mailer.Send to be called")
	}
}

func TestService_Send_envelopeAddresses(t *testing.T) {
	mailer := &spyMailer{}
	svc := newService(mailer)
	form := validForm()

	if err := svc.Send(context.Background(), form); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if mailer.mail.From != "site@example.com" {
		t.Errorf("From = %q, want %q", mailer.mail.From, "site@example.com")
	}
	if !slices.Contains(mailer.mail.To, "inbox@example.com") {
		t.Errorf("To = %v, want it to contain %q", mailer.mail.To, "inbox@example.com")
	}
	if mailer.mail.ReplyTo != form.Email {
		t.Errorf("ReplyTo = %q, want %q (so hitting reply answers the sender)", mailer.mail.ReplyTo, form.Email)
	}
}

func TestService_Send_subjectAndBody(t *testing.T) {
	mailer := &spyMailer{}
	svc := newService(mailer)
	form := validForm()

	if err := svc.Send(context.Background(), form); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if !strings.Contains(mailer.mail.Subject, form.Subject) {
		t.Errorf("Subject = %q, want it to contain %q", mailer.mail.Subject, form.Subject)
	}

	for _, want := range []string{form.Name, form.Email, form.Message} {
		if !strings.Contains(mailer.mail.Body, want) {
			t.Errorf("Body missing %q\n---\n%s", want, mailer.mail.Body)
		}
	}
}

func TestService_Send_validation(t *testing.T) {
	cases := []struct {
		name string
		form contact.Form
	}{
		{"missing name", contact.Form{Email: "a@b.c", Subject: "s", Message: "m"}},
		{"missing email", contact.Form{Name: "n", Subject: "s", Message: "m"}},
		{"missing subject", contact.Form{Name: "n", Email: "a@b.c", Message: "m"}},
		{"missing message", contact.Form{Name: "n", Email: "a@b.c", Subject: "s"}},
		{"email without @", contact.Form{Name: "n", Email: "no-at-sign", Subject: "s", Message: "m"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mailer := &spyMailer{}
			svc := newService(mailer)

			err := svc.Send(context.Background(), tc.form)
			if !errors.Is(err, contact.ErrInvalidForm) {
				t.Errorf("err = %v, want errors.Is(err, ErrInvalidForm)", err)
			}
			if mailer.called {
				t.Error("Mailer.Send must not be called for invalid forms")
			}
		})
	}
}

func TestService_Send_mailerErrorPropagates(t *testing.T) {
	boom := errors.New("smtp went boom")
	mailer := &spyMailer{err: boom}
	svc := newService(mailer)

	err := svc.Send(context.Background(), validForm())
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want errors.Is(err, boom)", err)
	}
}
