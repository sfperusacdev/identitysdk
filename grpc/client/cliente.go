package client

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/configs"
	identitygrpc "github.com/sfperusacdev/identitysdk/grpc"
	"github.com/sfperusacdev/identitysdk/xreq"
	"go.uber.org/fx"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const grpcReadyTimeout = 5 * time.Second

var errGrpcConnShutdown = errors.New("grpc connection is shutdown")

type GrpcClient struct {
	config configs.GeneralServiceConfigProvider

	mu           sync.RWMutex
	conns        map[string]*gogrpc.ClientConn
	resourceCode string
}

var _ gogrpc.ClientConnInterface = (*GrpcClient)(nil)

func NewGrpcClient(lc fx.Lifecycle, resourceCode string, config configs.GeneralServiceConfigProvider) *GrpcClient {
	client := &GrpcClient{
		resourceCode: resourceCode,
		config:       config,
		conns:        make(map[string]*gogrpc.ClientConn),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.close()
		},
	})

	return client
}

// Invoke implements [gogrpc.ClientConnInterface].
func (g *GrpcClient) Invoke(ctx context.Context, method string, args any, reply any, opts ...gogrpc.CallOption) error {
	conn, err := g.Connection(ctx, g.resourceCode)
	if err != nil {
		return err
	}

	return conn.Invoke(ctx, method, args, reply, opts...)
}

// NewStream implements [gogrpc.ClientConnInterface].
func (g *GrpcClient) NewStream(ctx context.Context, desc *gogrpc.StreamDesc, method string, opts ...gogrpc.CallOption) (gogrpc.ClientStream, error) {
	conn, err := g.Connection(ctx, g.resourceCode)
	if err != nil {
		return nil, err
	}

	return conn.NewStream(ctx, desc, method, opts...)
}

func (g *GrpcClient) Connection(ctx context.Context, resourceCode string) (*gogrpc.ClientConn, error) {
	grpcURL, err := g.requestGrpcLocation(ctx, resourceCode)
	if err != nil {
		return nil, err
	}

	g.mu.RLock()
	conn := g.conns[grpcURL]
	g.mu.RUnlock()

	if conn != nil {
		readyConn, err := g.ensureConnection(ctx, resourceCode, grpcURL, conn)
		if err == nil {
			return readyConn, nil
		}

		if !errors.Is(err, errGrpcConnShutdown) {
			return nil, err
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if conn := g.conns[grpcURL]; conn != nil {
		if conn.GetState() != connectivity.Shutdown {
			readyConn, err := g.ensureConnection(ctx, resourceCode, grpcURL, conn)
			if err == nil {
				return readyConn, nil
			}

			if !errors.Is(err, errGrpcConnShutdown) {
				return nil, err
			}
		}

		delete(g.conns, grpcURL)
		if err := conn.Close(); err != nil {
			slog.Warn("failed to close shutdown grpc connection", "empresa", identitysdk.Empresa(ctx), "resource_code", resourceCode, "grpc_url", grpcURL, "error", err)
		}
	}

	conn, err = gogrpc.NewClient(
		grpcURL,
		gogrpc.WithTransportCredentials(insecure.NewCredentials()),
		gogrpc.WithUnaryInterceptor(g.unaryContextInterceptor),
		gogrpc.WithStreamInterceptor(g.streamContextInterceptor),
		gogrpc.WithConnectParams(gogrpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  200 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   5 * time.Second,
			},
			MinConnectTimeout: grpcReadyTimeout,
		}),
	)
	if err != nil {
		slog.Warn("failed to create grpc client", "empresa", identitysdk.Empresa(ctx), "resource_code", resourceCode, "grpc_url", grpcURL, "error", err)
		return nil, err
	}

	g.conns[grpcURL] = conn

	return g.ensureConnection(ctx, resourceCode, grpcURL, conn)
}

func (g *GrpcClient) unaryContextInterceptor(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *gogrpc.ClientConn,
	invoker gogrpc.UnaryInvoker,
	opts ...gogrpc.CallOption,
) error {
	return invoker(g.outgoingContext(ctx), method, req, reply, cc, opts...)
}

func (g *GrpcClient) streamContextInterceptor(
	ctx context.Context,
	desc *gogrpc.StreamDesc,
	cc *gogrpc.ClientConn,
	method string,
	streamer gogrpc.Streamer,
	opts ...gogrpc.CallOption,
) (gogrpc.ClientStream, error) {
	return streamer(g.outgoingContext(ctx), desc, cc, method, opts...)
}

func (g *GrpcClient) outgoingContext(ctx context.Context) context.Context {
	pairs := make([]string, 0, 12)
	appendPair := func(key, value string) {
		value = strings.TrimSpace(value)
		if value == "" || strings.Contains(value, "####") {
			return
		}
		pairs = append(pairs, key, value)
	}

	appendPair(identitygrpc.MetadataAccessToken, g.config.IdentityAccessToken())
	appendPair(identitygrpc.MetadataToken, identitysdk.Token(ctx))
	appendPair(identitygrpc.MetadataEmpresa, identitysdk.Empresa(ctx))
	appendPair(identitygrpc.MetadataUsername, identitysdk.Username(ctx))
	appendPair(identitygrpc.MetadataRequestOrigin, identitysdk.RequestOrigin(ctx))

	_, sucursal := identitysdk.Empresa_Sucursal(ctx)
	appendPair(identitygrpc.MetadataSucursal, sucursal)

	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

func (g *GrpcClient) ensureConnection(ctx context.Context, resourceCode string, grpcURL string, conn *gogrpc.ClientConn) (*gogrpc.ClientConn, error) {
	state := conn.GetState()

	if state == connectivity.Shutdown {
		g.mu.Lock()
		if current := g.conns[grpcURL]; current == conn {
			delete(g.conns, grpcURL)
		}
		g.mu.Unlock()

		if err := conn.Close(); err != nil {
			slog.Warn("failed to close shutdown grpc connection", "empresa", identitysdk.Empresa(ctx), "resource_code", resourceCode, "grpc_url", grpcURL, "error", err)
		}

		return nil, errGrpcConnShutdown
	}

	if state == connectivity.Idle || state == connectivity.TransientFailure {
		conn.Connect()
	}

	if err := waitForReady(ctx, conn); err != nil {
		slog.Warn("grpc connection is not ready", "empresa", identitysdk.Empresa(ctx), "resource_code", resourceCode, "grpc_url", grpcURL, "state", conn.GetState(), "error", err)
		return nil, err
	}

	return conn, nil
}

func waitForReady(ctx context.Context, conn *gogrpc.ClientConn) error {
	waitCtx := ctx
	cancel := func() {}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		waitCtx, cancel = context.WithTimeout(ctx, grpcReadyTimeout)
	}
	defer cancel()

	for {
		state := conn.GetState()

		if state == connectivity.Ready {
			return nil
		}

		if state == connectivity.Shutdown {
			return errGrpcConnShutdown
		}

		if !conn.WaitForStateChange(waitCtx, state) {
			if err := waitCtx.Err(); err != nil {
				return err
			}

			return errors.New("grpc connection state did not change")
		}
	}
}

func (g *GrpcClient) close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	var closeErr error

	for grpcURL, conn := range g.conns {
		if err := conn.Close(); err != nil {
			slog.Warn("failed to close grpc connection", "grpc_url", grpcURL, "error", err)

			if closeErr == nil {
				closeErr = err
			}
		}

		delete(g.conns, grpcURL)
	}

	return closeErr
}

func (g *GrpcClient) requestGrpcLocation(ctx context.Context, resourceCode string) (string, error) {
	accessToken := g.config.IdentityAccessToken()
	companyCode := identitysdk.Empresa(ctx)

	var apiResponse struct {
		Message string `json:"message"`
		Data    struct {
			Location string `json:"location"`
		} `json:"data"`
	}

	if err := xreq.MakeRequest(
		ctx,
		g.config.Identity(),
		"/api/v1/get-service-grpc-location",
		xreq.WithJsonContentType(),
		xreq.WithQueryParam("company_code", companyCode),
		xreq.WithQueryParam("resource_code", resourceCode),
		xreq.WithAccessToken(accessToken),
		xreq.WithUnmarshalResponseInto(&apiResponse),
	); err != nil {
		slog.Warn("failed to request grpc location", "empresa", companyCode, "resource_code", resourceCode, "error", err)
		return "", err
	}

	location := strings.TrimSpace(apiResponse.Data.Location)
	if location == "" {
		err := errors.New("grpc location is empty")
		slog.Warn("invalid grpc location response", "empresa", companyCode, "resource_code", resourceCode, "message", apiResponse.Message)
		return "", err
	}

	return location, nil
}
