package grpcapp

import (
	"fmt"
	"github.com/leandoer69/golang-sso-grpc/internal/grpc/auth"
	authgrpc "github.com/leandoer69/golang-sso-grpc/internal/grpc/auth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	port int,
) *App {

	server := grpc.NewServer()
	auth.Register(server, authService)

	return &App{
		log:        log,
		grpcServer: server,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic("failed to run grpc app: " + err.Error())
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	a.log.Info("starting gRPC server")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := a.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("gRPC server is running", slog.String("addr", lis.Addr().String()))
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op))
	a.log.Info("stopping grpc server")

	a.grpcServer.GracefulStop()
}
