package sms

import "context"

type SmsSenderResponse struct {
	TxId string
}
type SmsSender interface {
	Send(ctx context.Context, number string, textMessage string) (*SmsSenderResponse, error)
}
