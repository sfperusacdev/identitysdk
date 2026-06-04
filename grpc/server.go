package grpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"

	"github.com/labstack/gommon/color"
	"github.com/sfperusacdev/identitysdk"
	"go.uber.org/fx"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ServiceTag = `group:"grpc-services"`

type ServeAddress string

type Service interface {
	Register(server gogrpc.ServiceRegistrar)
}

var Module = fx.Module("grpc-server",
	fx.Provide(
		fx.Annotate(
			NewServer,
			fx.ParamTags(ServiceTag),
		),
	),
)

func AsService(fn any) any {
	return fx.Annotate(
		fn,
		fx.As(new(Service)),
		fx.ResultTags(ServiceTag),
	)
}

func NewServer(services []Service) *gogrpc.Server {
	server := gogrpc.NewServer(
		gogrpc.UnaryInterceptor(apiKeyUnaryInterceptor),
		gogrpc.StreamInterceptor(apiKeyStreamInterceptor),
	)
	for _, service := range services {
		service.Register(server)
	}
	return server
}

func apiKeyUnaryInterceptor(
	ctx context.Context,
	req any,
	info *gogrpc.UnaryServerInfo,
	handler gogrpc.UnaryHandler,
) (any, error) {
	ctx, err := contextWithAPIKeyMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func apiKeyStreamInterceptor(
	srv any,
	stream gogrpc.ServerStream,
	info *gogrpc.StreamServerInfo,
	handler gogrpc.StreamHandler,
) error {
	ctx, err := contextWithAPIKeyMetadata(stream.Context())
	if err != nil {
		return err
	}
	return handler(srv, &contextServerStream{ServerStream: stream, ctx: ctx})
}

type contextServerStream struct {
	gogrpc.ServerStream
	ctx context.Context
}

func (s *contextServerStream) Context() context.Context {
	return s.ctx
}

func contextWithAPIKeyMetadata(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata no encontrada")
	}

	apikey := firstMetadataValue(md, MetadataAPIKey)
	if apikey == "" {
		return nil, status.Error(codes.Unauthenticated, "API KEY no encontrado")
	}

	data, err := identitysdk.ValidateApiKeyWithCache(ctx, apikey)
	if err != nil {
		slog.Warn("gRPC API key validation failed", "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if data == nil {
		return nil, status.Error(codes.Unauthenticated, "api key session invalida")
	}

	newCtx := identitysdk.BuildApikeyContext(ctx, apikey, &data.Apikey)

	if empresa := firstMetadataValue(md, MetadataEmpresa); empresa != "" {
		newCtx = identitysdk.CtxWithDomain(newCtx, empresa)
	}
	if token := firstMetadataValue(md, MetadataToken); token != "" {
		newCtx = identitysdk.CtxWithToken(newCtx, token)
	}
	if username := firstMetadataValue(md, MetadataUsername); username != "" {
		newCtx = identitysdk.CtxWithUsername(newCtx, username)
	}
	if sucursal := firstMetadataValue(md, MetadataSucursal); sucursal != "" {
		newCtx = identitysdk.CtxWithSucursal(newCtx, sucursal)
	}
	if origin := firstMetadataValue(md, MetadataRequestOrigin); origin != "" {
		newCtx = identitysdk.CtxWithRequestOrigin(newCtx, origin)
	}

	return newCtx, nil
}

func firstMetadataValue(md metadata.MD, key string) string {
	for _, value := range md.Get(key) {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func StartServer(lc fx.Lifecycle, server *gogrpc.Server, address ServeAddress) {
	colorer := color.New()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			address := strings.TrimSpace(string(address))
			if address == "" {
				slog.Info("gRPC server skipped: address is empty")
				return nil
			}

			listener, err := net.Listen("tcp", address)
			if err != nil {
				slog.Error("failed to listen gRPC server", "address", address, "error", err)
				return err
			}

			go func() {
				colorer.Printf("⇨ gRPC server started on %s\n", colorer.Green(listener.Addr()))
				if err := server.Serve(listener); err != nil && !errors.Is(err, gogrpc.ErrServerStopped) {
					slog.Error("gRPC server error", "error", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Shutting down gRPC server")
			server.GracefulStop()
			slog.Info("gRPC server stopped successfully")
			return nil
		},
	})
}
