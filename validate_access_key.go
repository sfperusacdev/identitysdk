package identitysdk

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/goones/errs"
)

func ValidateAccessKey(ctx context.Context, access_token string) error {
	var buff bytes.Buffer

	if err := json.NewEncoder(&buff).
		Encode(map[string]string{
			"access_token": access_token,
		}); err != nil {
		slog.Error("failed to encode API key payload", "error", err)
		return errs.InternalErrorDirect(errs.ErrInternal)
	}
	return xreq.MakeRequest(ctx,
		identityAddress,
		"/v1/check-access-token",
		xreq.WithMethod(http.MethodPost),
		xreq.WithJsonContentType(),
		xreq.WithRequestBody(&buff),
	)
}
