package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/lostmyescape/link-shortener/common/kafka"
	"github.com/lostmyescape/link-shortener/common/logger/sl"
	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
	"github.com/lostmyescape/link-shortener/sso/internal/storage"
	"github.com/lostmyescape/link-shortener/sso/pkg/jwt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log                *slog.Logger
	usrSaver           UserSaver
	usrProvider        UserProvider
	appProvider        AppProvider
	tokenTTL           time.Duration
	rTokenTTL          time.Duration
	tokenStoreProvider TokenStoreProvider
	producerProvider   ProducerProvider
	ip                 string
}

type TokenStoreProvider interface {
	SaveToken(ctx context.Context, userID int64, token string, ttl time.Duration) error
	DeleteToken(ctx context.Context, userID int64) error
	GetToken(ctx context.Context, userID int64) (string, error)
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type ProducerProvider interface {
	Publish(ctx context.Context, key string, value interface{}) error
	Close() error
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidAppId        = errors.New("invalid app ID")
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// New returns a new instance of the Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
	rTokenTTL time.Duration,
	tokenStoreProvider TokenStoreProvider,
	producerProvider ProducerProvider,
	ip string,
) *Auth {
	return &Auth{
		log:                log,
		usrSaver:           userSaver,
		usrProvider:        userProvider,
		appProvider:        appProvider,
		tokenTTL:           tokenTTL,
		rTokenTTL:          rTokenTTL,
		tokenStoreProvider: tokenStoreProvider,
		producerProvider:   producerProvider,
		ip:                 ip,
	}
}

// Login checks if user with given credentials exists in the system
//
// If password is incorrect, returns error
// If user doesn't exist, returns error
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, string, error) {
	const op = "auth.LoginUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	rToken, err := jwt.NewToken(user, app, a.rTokenTTL)
	if err != nil {
		a.log.Error("failed to generate refresh token", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if a.tokenStoreProvider == nil {
		a.log.Error("token store is nil, cannot save token")
		return "", "", fmt.Errorf("%s: %w", op, fmt.Errorf("token store not initialized"))
	}

	if err := a.tokenStoreProvider.SaveToken(ctx, user.ID, rToken, a.rTokenTTL); err != nil {
		a.log.Error("failed to save token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tokens successfully generated")

	ev := map[string]interface{}{
		"type":      kafka.EventUserLoggedIn,
		"timestamp": time.Now().UTC(),
		"user_id":   user.ID,
		"email":     user.Email,
		"ip":        a.ip,
	}

	log.Info("user logged successfully")

	err = a.producerProvider.Publish(ctx, strconv.FormatInt(user.ID, 10), ev)
	if err != nil {
		a.log.Error("failed to send message to Kafka", sl.Err(err))
	}

	return token, rToken, nil
}

func (a *Auth) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (string, string, error) {
	const op = "auth.RefreshToken"

	log := a.log.With(
		slog.String("op", op),
	)

	app, err := a.appProvider.App(ctx, 1)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	user, err := jwt.ParseToken(refreshToken, app.Secret)
	if err != nil {
		a.log.Error("invalid token or failed to parse token", sl.Err(err))
		return "", "", ErrInvalidRefreshToken
	}

	storedToken, err := a.tokenStoreProvider.GetToken(ctx, user.ID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.log.Error("token expired or invalid", sl.Err(err))
			return "", "", fmt.Errorf("token expired or invalid: %w", err)
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if storedToken != refreshToken {
		a.log.Error("failed to compare tokens")
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// delete the old token
	err = a.tokenStoreProvider.DeleteToken(ctx, user.ID)
	if err != nil {
		a.log.Error("failed to delete token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("old token deleted")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	rToken, err := jwt.NewToken(user, app, a.rTokenTTL)
	if err != nil {
		a.log.Error("failed to generate rToken", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = a.tokenStoreProvider.SaveToken(ctx, user.ID, rToken, a.rTokenTTL)
	if err != nil {
		a.log.Error("failed to save token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tokens successfully generated")

	return token, rToken, nil
}

// Register registers a new user in the system and returns a user ID
// If user with given username already exists, returns error
func (a *Auth) Register(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.Register"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, hashedPass)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	ev := map[string]interface{}{
		"type":      kafka.EventUserRegistered,
		"timestamp": time.Now().UTC(),
		"user_id":   id,
		"email":     email,
		"ip":        a.ip,
	}

	err = a.producerProvider.Publish(ctx, strconv.FormatInt(id, 10), ev)
	if err != nil {
		a.log.Error("failed to send message to Kafka", sl.Err(err))
	}

	log.Info("user registered")

	return id, nil
}

// IsAdmin checks if user is admin
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppId)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checking if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) Logout(ctx context.Context, token string) (string, error) {
	const op = "sso.internal.services.auth.Logout"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("user logout attempt...")

	app, err := a.appProvider.App(ctx, 1)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	user, err := jwt.ParseToken(token, app.Secret)
	if err != nil {
		a.log.Error("failed to parse token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = a.tokenStoreProvider.DeleteToken(ctx, user.ID)
	if err != nil {
		a.log.Error("failed to delete token")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	ev := map[string]interface{}{
		"type":      kafka.EventUserLoggedOut,
		"timestamp": time.Now().UTC(),
		"user_id":   user.ID,
		"email":     user.Email,
		"ip":        a.ip,
	}

	err = a.producerProvider.Publish(ctx, strconv.FormatInt(user.ID, 10), ev)
	if err != nil {
		a.log.Error("failed to send message to Kafka", sl.Err(err))
	}

	log.Info("refresh token was deleted")

	return "successful logout", nil
}
