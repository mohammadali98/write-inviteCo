package notification

import (
	"context"
	"fmt"
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

func (s *ResendSender) SendOrderConfirmationEmail(ctx context.Context, customerEmail string, orderID int64) error {
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

	email := strings.TrimSpace(customerEmail)
	if email == "" {
		return fmt.Errorf("customer email is empty")
	}

	from := s.fromEmail
	if !strings.Contains(from, "<") {
		from = fmt.Sprintf("Write&InviteCo <%s>", from)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{email},
		Subject: fmt.Sprintf("Order Confirmed #%d", orderID),
		Html: fmt.Sprintf(
			"<strong>Your order #%d is confirmed.</strong><p>Thank you for choosing Write&InviteCo.</p>",
			orderID,
		),
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		log.Println("RESEND ERROR:", err)
		return err
	}

	log.Println("EMAIL SENT SUCCESS:", sent.Id)
	return nil
}
