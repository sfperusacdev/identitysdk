package email

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/resend/resend-go/v2"
)

type ResendEmail struct {
	client       *resend.Client
	outgoingMail string
}

var _ EmailSender = (*ResendEmail)(nil)
var emailRgx = regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})$`)

func NewResendEmail(apiKey string, outgoingMail string) (*ResendEmail, error) {
	if !(emailRgx.MatchString(outgoingMail) || outgoingMail == "Acme <onboarding@resend.dev>") {
		err := fmt.Errorf("invalid outgoingMail: %s", outgoingMail)
		slog.Error("failed to create ResendEmail", "error", err)
		return nil, err
	}
	client := resend.NewClient(apiKey)
	return &ResendEmail{client: client, outgoingMail: outgoingMail}, nil
}

func (re *ResendEmail) Send(ctx context.Context, to string, subject string, htmlContent string) (EmailSenderResponse, error) {
	params := resend.SendEmailRequest{
		From:    re.outgoingMail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlContent,
	}

	res, err := re.client.Emails.Send(&params)
	if err != nil {
		slog.Warn("failed to send email", "to", to, "subject", subject, "error", err)
		return EmailSenderResponse{}, fmt.Errorf("failed to send email: %w", err)
	}
	return EmailSenderResponse{TxId: res.Id}, nil
}
