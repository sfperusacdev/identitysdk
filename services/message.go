package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/services/internal/sms"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type SendEmailDetails struct {
	TxId string
}

func (s *ExternalBridgeService) SendEmail(ctx context.Context, to string, subject string, htmlContent string, tags ...string) (*SendEmailDetails, error) {
	domain := identitysdk.Empresa(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return nil, err
	}

	var id = uuid.NewString()
	var data = map[string]string{
		"id":              id,
		"company_code":    domain,
		"recipient_email": to,
		"subject":         subject,
		"body":            htmlContent,
		"tags":            strings.Join(tags, ", "),
	}

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	if err := encoder.Encode(data); err != nil {
		slog.Error("Failed to encode email data to JSON", "data", data, "error", err)
		return nil, err
	}

	if err := xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/push/email",
		xreq.WithJsonContentType(),
		xreq.WithRequestBody(&buff),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	); err != nil {
		slog.Error("Failed to send email request", "baseurl", baseurl, "endpoint", "/api/v1/_internal/push/email", "error", err)
		return nil, err
	}

	return &SendEmailDetails{TxId: id}, nil
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
