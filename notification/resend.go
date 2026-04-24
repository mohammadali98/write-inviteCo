package notification

import (
	"context"
	"fmt"
	"html"
	"log"
	"strings"

	resend "github.com/resend/resend-go/v2"
)

type ResendSender struct {
	apiKey    string
	fromEmail string
	client    *resend.Client
}

func NewResendSender(apiKey string, fromEmail string) *ResendSender {
	key := strings.TrimSpace(apiKey)
	return &ResendSender{
		apiKey:    key,
		fromEmail: strings.TrimSpace(fromEmail),
		client:    resend.NewClient(key),
	}
}

// NOTE:
// In Resend free/testing mode, emails only go to verified email.
// To send to real users:
// - Verify domain in Resend dashboard
// - Use domain email like: orders@yourdomain.com
func (s *ResendSender) SendOrderEmail(ctx context.Context, to string, subject string, body string) error {
	return s.sendOrderEmail(ctx, to, subject, body)
}

func (s *ResendSender) sendOrderEmail(ctx context.Context, to string, subject string, body string) error {
	if s.apiKey == "" {
		return fmt.Errorf("resend api key is not configured")
	}
	if s.fromEmail == "" {
		return fmt.Errorf("resend from email is not configured")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	email := strings.TrimSpace(to)
	if email == "" {
		return fmt.Errorf("recipient email is empty")
	}

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("email subject is empty")
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("email body is empty")
	}

	from := s.fromEmail
	if !strings.Contains(from, "<") {
		from = fmt.Sprintf("Write&InviteCo <%s>", from)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{email},
		Subject: subject,
		Text:    body,
		Html: fmt.Sprintf(
			"<div style=\"font-family:Arial,sans-serif;white-space:pre-wrap;line-height:1.6;\">%s</div>",
			html.EscapeString(body),
		),
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		log.Println("Email send failed:", err)
		return err
	}

	log.Println("Email sent successfully:", sent.Id)
	return nil
}
