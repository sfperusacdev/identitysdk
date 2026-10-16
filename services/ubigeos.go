package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/sark_services/global"
	"github.com/sfperusacdev/identitysdk/xreq"
)

// Deprecated: use global.Ubigeo instead.
type Ubigeo = global.Ubigeo

// Deprecated: use bridge.Global.GetUbigeos instead.
func (s *ExternalBridgeService) GetUbigeos(ctx context.Context) ([]Ubigeo, error) {
	company, token := s.readCompanyAndToken(ctx)
	baseURL, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}

	var response struct {
		Message string   `json:"message"`
		Data    []Ubigeo `json:"data"`
	}

	if err := xreq.MakeRequest(ctx,
		baseURL, "/api/v2/ubigeos",
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// Deprecated: use global.UbigeoSource instead.
type UbigeoSource = global.UbigeoSource

const (
	// Deprecated: use global.UbigeoSourceINEI instead.
	UbigeoSourceINEI = global.UbigeoSourceINEI
	// Deprecated: use global.UbigeoSourceRENIEC instead.
	UbigeoSourceRENIEC = global.UbigeoSourceRENIEC
	// Deprecated: use global.UbigeoSourceSUNAT instead.
	UbigeoSourceSUNAT = global.UbigeoSourceSUNAT
)

// Deprecated: use bridge.Global.SearchUbigeo instead.
func (s *ExternalBridgeService) SearchUbigeo(
	ctx context.Context,
	source UbigeoSource,
	departmentName string,
	provinceName string,
	districtName string,
) (*Ubigeo, error) {
	company, token := s.readCompanyAndToken(ctx)

	serviceURL, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving general service URL", "error", err)
		return nil, err
	}

	var result struct {
		Message string  `json:"message"`
		Data    *Ubigeo `json:"data"`
	}

	route := fmt.Sprintf(
		"/api/v2/ubigeos/%s/%s/%s/%s",
		string(source),
		departmentName,
		provinceName,
		districtName,
	)

	if err := xreq.MakeRequest(
		ctx,
		serviceURL,
		route,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&result),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}

	return result.Data, nil
}
