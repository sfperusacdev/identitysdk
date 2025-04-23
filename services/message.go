package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/services/internal/email"
	"github.com/sfperusacdev/identitysdk/services/internal/sms"
)

func (s *ExternalBridgeService) SendEmail(ctx context.Context, to string, subject string, htmlContent string) (*email.EmailSenderResponse, error) {
	domain := identitysdk.Empresa(ctx)

	apikey, err := s.ReadVariable(ctx, "RESEND_API_KEY")
	if err != nil {
		slog.Error("failed to read RESEND_API_KEY", "domain", domain, "error", err)
		return nil, fmt.Errorf("reading RESEND_API_KEY: %w", err)
	}

	outgoingMail, err := s.ReadVariable(ctx, "RESEND_OUTGOING_MAIL")
	if err != nil {
		slog.Error("failed to read RESEND_OUTGOING_MAIL", "domain", domain, "error", err)
		return nil, fmt.Errorf("reading RESEND_OUTGOING_MAIL: %w", err)
	}

	client, err := email.NewResendEmail(apikey, outgoingMail)
	if err != nil {
		slog.Error("failed to initialize ResendEmail client", "domain", domain, "error", err)
		return nil, fmt.Errorf("initializing email client: %w", err)
	}

	res, err := client.Send(ctx, to, subject, htmlContent)
	if err != nil {
		slog.Error("failed to send email", "domain", domain, "to", to, "subject", subject, "error", err)
		return nil, fmt.Errorf("sending email: %w", err)
	}
	return &res, nil
}

// region SMS

func (s *ExternalBridgeService) SendSMS(ctx context.Context, number string, textMessage string) (*sms.SmsSenderResponse, error) {
	domain := identitysdk.Empresa(ctx)

	username, err := s.ReadVariable(ctx, "LABSMOBILE_USERNAME")
	if err != nil {
		slog.Error("failed to read LABSMOBILE_USERNAME", "domain", domain, "error", err)
		return nil, fmt.Errorf("reading LABSMOBILE_USERNAME: %w", err)
	}

	password, err := s.ReadVariable(ctx, "LABSMOBILE_PASSWORD")
	if err != nil {
		slog.Error("failed to read LABSMOBILE_PASSWORD", "domain", domain, "error", err)
		return nil, fmt.Errorf("reading LABSMOBILE_PASSWORD: %w", err)
	}

	client, err := sms.NewLabsmobile(username, password)
	if err != nil {
		slog.Error("failed to initialize Labsmobile client", "domain", domain, "error", err)
		return nil, fmt.Errorf("initializing Labsmobile client: %w", err)
	}

	res, err := client.Send(ctx, number, textMessage)
	if err != nil {
		slog.Error("failed to send SMS", "domain", domain, "number", number, "error", err)
		return nil, fmt.Errorf("sending SMS: %w", err)
	}

	slog.Info("SMS sent successfully", "domain", domain, "number", number, "tx_id", res.TxId)
	return res, nil
}
