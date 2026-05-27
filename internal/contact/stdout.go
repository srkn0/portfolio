package contact

import (
	"context"
	"fmt"
)

type StdoutMailer struct{}

func (StdoutMailer) Send(_ context.Context, m Mail) error {
	fmt.Printf("--- MAIL ---\nFrom: %s\nTo: %v\nReply-To: %s\nSubject: %s\n\n%s\n--- END ---\n",
		m.From, m.To, m.ReplyTo, m.Subject, m.Body)
	return nil
}
