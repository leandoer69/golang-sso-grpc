package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/leandoer69/golang-sso-grpc/internal/domain/models"
	"github.com/leandoer69/golang-sso-grpc/internal/lib/jwt"
	"github.com/leandoer69/golang-sso-grpc/internal/lib/logger/sl"
	"github.com/leandoer69/golang-sso-grpc/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int32,
) (string, error) {
	const op = "auth.Login"

	a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	a.log.Info("attempt to log in")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", sl.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials")

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("app not found", sl.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		a.log.Error("failed to get app", sl.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("user logged in successfully")

	token, err := jwt.NewToken(app, user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.RegisterNewUser"

	a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	a.log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user already exists", sl.Error(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		a.log.Error("failed to save user", sl.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("user was registered")
	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	a.log.Info("checking if user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.With("user not found", sl.Error(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.With("failed to get user", sl.Error(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
