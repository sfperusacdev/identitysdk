package identitysdk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func IdentityServerCheckHealth() error {
	hostUrl, err := url.JoinPath(identityAddress, "/health")
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, hostUrl, nil)
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(res.Body)
		return fmt.Errorf("identity server error: %s", string(payload))
	}
	return nil
}
