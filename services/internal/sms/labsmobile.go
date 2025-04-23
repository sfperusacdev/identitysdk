package sms

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type Labsmobile struct {
	labsmobileUrl string
	authorization string
}

var _ SmsSender = (*Labsmobile)(nil)

func NewLabsmobile(username string, password string) (*Labsmobile, error) {
	var value = fmt.Sprintf("%s:%s", username, password)
	var token = base64.StdEncoding.EncodeToString([]byte(value))
	var authorization = fmt.Sprintf("Basic %s", token)

	return &Labsmobile{
		authorization: authorization,
		labsmobileUrl: `https://api.labsmobile.com/json/send`,
	}, nil
}

func (l *Labsmobile) Send(ctx context.Context, number string, textMessage string) (*SmsSenderResponse, error) {
	params := map[string]any{
		"message": textMessage,
		"tpoa":    "Sender",
		"recipient": []map[string]string{
			{"msisdn": number},
		},
	}

	jsonData, err := json.Marshal(&params)
	if err != nil {
		slog.Error("failed to marshal SMS params", "number", number, "error", err)
		return nil, fmt.Errorf("marshal SMS params: %w", err)
	}

	requestBody := bytes.NewReader(jsonData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.labsmobileUrl, requestBody)
	if err != nil {
		slog.Error("failed to create SMS request", "number", number, "error", err)
		return nil, fmt.Errorf("create SMS request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", l.authorization)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send SMS request", "number", number, "error", err)
		return nil, fmt.Errorf("send SMS request: %w", err)
	}
	defer res.Body.Close()

	var responseObject struct {
		Subid   string `json:"subid"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(res.Body).Decode(&responseObject); err != nil {
		slog.Error("failed to decode SMS response", "number", number, "error", err)
		return nil, fmt.Errorf("decode SMS response: %w", err)
	}

	if strings.Contains(responseObject.Message, "successfully") {
		return &SmsSenderResponse{TxId: responseObject.Subid}, nil
	}
	slog.Error("SMS failed", "number", number, "code", responseObject.Code, "message", responseObject.Message)
	return nil, errors.New(responseObject.Message)
}
