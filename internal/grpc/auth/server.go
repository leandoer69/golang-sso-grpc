package auth

import (
	"context"
	"errors"
	"github.com/leandoer69/golang-sso-grpc/internal/services/auth"
	"github.com/leandoer69/golang-sso-grpc/internal/storage"
	ssov1 "github.com/leandoer69/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appID int32,
	) (token string, err error)
	RegisterNewUser(ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(grpcServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(grpcServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context,
	req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {

	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		if errors.Is(err, auth.ErrInvalidAppID) {
			return nil, status.Error(codes.InvalidArgument, "invalid app_id")
		}

		return nil, status.Error(codes.Internal, "failed to log in")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func (s *serverAPI) Register(ctx context.Context,
	req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {

	if err := validateRegister(req); err != nil {
		return nil, err
	}

	uid, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register new user")
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

func (s *serverAPI) IsAdmin(ctx context.Context,
	req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to check admin status")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	return nil
}
