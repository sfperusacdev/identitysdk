package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type SendDetails struct {
	TxId string
}

type Mail struct {
	CompanyCode    string   `json:"company_code"`
	EecipientEmail string   `json:"recipient_email"`
	Subject        string   `json:"subject"`
	Body           string   `json:"body"`
	Tags           []string `json:"tags"`
}

func (s *ExternalBridgeService) SendBatchMails(ctx context.Context, mails ...Mail) error {
	if len(mails) == 0 {
		return nil
	}

	domain := identitysdk.Empresa(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return err
	}

	var elements = make([]map[string]string, 0, len(mails))
	for _, m := range mails {
		elements = append(elements, map[string]string{
			"id":              uuid.NewString(),
			"company_code":    domain,
			"recipient_email": m.EecipientEmail,
			"subject":         m.Subject,
			"body":            m.Body,
			"tags":            strings.Join(m.Tags, ";"),
		})
	}
	var data = map[string]any{
		"skip_errors": true,
		"mails":       elements,
	}

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	if err := encoder.Encode(data); err != nil {
		slog.Error("Failed to encode mails data to JSON",
			"domain", domain,
			"data", data,
			"error", err,
		)
		return err
	}
	return xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/push/email",
		xreq.WithJsonContentType(),
		xreq.WithRequestBody(&buff),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	)
}

func (s *ExternalBridgeService) SendMail(ctx context.Context, to string, subject string, htmlContent string, tags ...string) (*SendDetails, error) {
	domain := identitysdk.Empresa(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return nil, err
	}

	var id = uuid.NewString()
	var data = map[string]any{
		"mails": map[string]string{
			"id":              id,
			"company_code":    domain,
			"recipient_email": to,
			"subject":         subject,
			"body":            htmlContent,
			"tags":            strings.Join(tags, ";"),
		},
	}

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	if err := encoder.Encode(data); err != nil {
		slog.Error("Failed to encode email data to JSON",
			"domain", domain,
			"data", data,
			"error", err,
		)
		return nil, err
	}

	if err := xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/push/email",
		xreq.WithJsonContentType(),
		xreq.WithRequestBody(&buff),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	); err != nil {
		slog.Error("Failed to send email request",
			"domain", domain,
			"baseurl", baseurl,
			"endpoint", "/api/v1/_internal/push/email",
			"error", err,
		)
		return nil, err
	}

	return &SendDetails{TxId: id}, nil
}

type SMS struct {
	CompanyCode    string   `json:"company_code"`
	RecipientPhone string   `json:"recipient_phone"`
	Message        string   `json:"message"`
	Tags           []string `json:"tags"`
}

// region sms
func (s *ExternalBridgeService) SendBatchSMS(ctx context.Context, smss ...SMS) error {
	if len(smss) == 0 {
		return nil
	}
	domain := identitysdk.Empresa(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return err
	}

	var elements = make([]map[string]string, 0, len(smss))
	for _, m := range smss {
		elements = append(elements, map[string]string{
			"id":              uuid.NewString(),
			"company_code":    domain,
			"recipient_phone": m.RecipientPhone,
			"message":         m.Message,
			"tags":            strings.Join(m.Tags, ";"),
		})
	}
	var data = map[string]any{
		"skip_errors": true,
		"sms":         elements,
	}

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	if err := encoder.Encode(data); err != nil {
		slog.Error("Failed to encode sms data to JSON",
			"domain", domain,
			"data", data,
			"error", err,
		)
		return err
	}
	return xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/push/sms",
		xreq.WithJsonContentType(),
		xreq.WithRequestBody(&buff),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	)
}

func (s *ExternalBridgeService) SendSMS(ctx context.Context, number string, message string, tags ...string) (*SendDetails, error) {
	domain := identitysdk.Empresa(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return nil, err
	}

	var id = uuid.NewString()
	var data = map[string]any{
		"sms": map[string]string{
			"id":              id,
			"company_code":    domain,
			"recipient_phone": number,
			"message":         message,
			"tags":            strings.Join(tags, ";"),
		},
	}

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	if err := encoder.Encode(data); err != nil {
		slog.Error("Failed to encode sms data to JSON",
			"domain", domain,
			"data", data,
			"error", err,
		)
		return nil, err
	}

	if err := xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/push/sms",
		xreq.WithJsonContentType(),
		xreq.WithRequestBody(&buff),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	); err != nil {
		slog.Error("Failed to send sms request",
			"domain", domain,
			"baseurl", baseurl,
			"endpoint", "/api/v1/_internal/push/sms",
			"error", err,
		)
		return nil, err
	}

	return &SendDetails{TxId: id}, nil
}

// region Retrive
type MailRecord struct {
	ID             uuid.UUID `json:"id"`
	CompanyCode    *string   `json:"company_code,omitempty"`
	RecipientEmail string    `json:"recipient_email"`

	Tags *string `json:"tags,omitempty"`

	Subject      string  `json:"subject"`
	Body         string  `json:"body"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message,omitempty"`
	RetryCount   int     `json:"retry_count"`

	RefId *string `json:"ref_id,omitempty"`

	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// GetEmailsByTag searches for emails that have been sent or are queued for sending.
// The search is filtered using the provided tag, which supports partial matching (e.g., "promo%").
// The filter applies to the "tags" field, using a LIKE query with a trailing wildcard for broader matches.
func (s *ExternalBridgeService) GetEmailsByTag(ctx context.Context, tagFilter ...string) ([]MailRecord, error) {
	filtered := []string{}
	for _, tag := range tagFilter {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		filtered = append(filtered, tag)
	}
	if len(filtered) == 0 {
		return []MailRecord{}, nil
	}

	domain := identitysdk.Empresa(ctx)
	token := identitysdk.Token(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return nil, err
	}
	queryParams := url.Values{}
	for _, tag := range filtered {
		queryParams.Add("q", tag)
	}
	var res struct {
		Data []MailRecord `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/retrive/email",
		xreq.WithAuthorization(token),
		xreq.WithQueryParams(queryParams),
		xreq.WithUnmarshalResponseInto(&res),
	); err != nil {
		slog.Error("Failed to retrieve Emails",
			"domain", domain,
			"baseurl", baseurl,
			"endpoint", "/api/v1/_internal/retrive/email",
			"error", err.Error(),
		)
		return nil, err
	}
	return res.Data, nil
}

type SMSRecord struct {
	ID             uuid.UUID `json:"id"`
	CompanyCode    *string   `json:"company_code,omitempty"`
	RecipientPhone string    `json:"recipient_phone"`
	Message        string    `json:"message"`

	Tags *string `json:"tags,omitempty"`

	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message,omitempty"`
	RetryCount   int     `json:"retry_count"`

	RefID *string `json:"ref_id,omitempty"`

	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// GetMessagesByTag searches for SMS messages that have been sent or are queued for sending.
// The search is filtered using the provided tag, which supports partial matching (e.g., "alert%").
// The filter applies to the "tags" field, using a LIKE query with a trailing wildcard for broader matches.
func (s *ExternalBridgeService) GetMessagesByTag(ctx context.Context, tagFilter ...string) ([]SMSRecord, error) {
	filtered := []string{}
	for _, tag := range tagFilter {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		filtered = append(filtered, tag)
	}
	if len(filtered) == 0 {
		return []SMSRecord{}, nil
	}

	domain := identitysdk.Empresa(ctx)
	token := identitysdk.Token(ctx)
	baseurl, err := identitysdk.GetMensajeriaServiceURL(ctx, domain)
	if err != nil {
		slog.Error("Failed to retrieve 'mensajeria' service URL", "domain", domain, "error", err)
		return nil, err
	}
	queryParams := url.Values{}
	for _, tag := range filtered {
		queryParams.Add("q", tag)
	}
	var res struct {
		Data []SMSRecord `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		baseurl, "/api/v1/_internal/retrive/sms",
		xreq.WithAuthorization(token),
		xreq.WithQueryParams(queryParams),
		xreq.WithUnmarshalResponseInto(&res),
	); err != nil {
		slog.Error("Failed to retrieve SMS",
			"domain", domain,
			"baseurl", baseurl,
			"endpoint", "/api/v1/_internal/retrive/sms",
			"error", err.Error(),
		)
		return nil, err
	}
	return res.Data, nil
}
