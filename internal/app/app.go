package app

import (
	grpcapp "github.com/leandoer69/golang-sso-grpc/internal/app/grpc"
	"github.com/leandoer69/golang-sso-grpc/internal/services/auth"
	"github.com/leandoer69/golang-sso-grpc/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
