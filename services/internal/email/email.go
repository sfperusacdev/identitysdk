package email

import "context"

type EmailSenderResponse struct {
	TxId string
}

type EmailSender interface {
	Send(ctx context.Context, to string, subject string, content string) (EmailSenderResponse, error)
}
