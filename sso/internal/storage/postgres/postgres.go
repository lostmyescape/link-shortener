package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/lostmyescape/link-shortener/sso/internal/config"
	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
	"github.com/lostmyescape/link-shortener/sso/internal/storage"
)

type Storage struct {
	DB *sql.DB
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Storage.Host,
		cfg.Storage.Port,
		cfg.Storage.User,
		cfg.Storage.Password,
		cfg.Storage.DbName,
		cfg.Storage.SslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connection failed to database: %w", err)
	}

	log.Println("Successful database connection")

	return &Storage{DB: db}, nil
}

// SaveUser creates a new user
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "postgres.SaveUser"

	var userID int64

	query := "INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id"

	err := s.DB.QueryRowContext(ctx, query, email, passHash).Scan(&userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				if pqErr.Constraint == "users_email_key" {
					return 0, storage.ErrUserAlreadyExists
				}
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	var user models.User

	err := s.DB.QueryRowContext(
		ctx,
		"SELECT id, email, pass_hash FROM users WHERE email = $1",
		email).
		Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	var isAdmin bool

	err := s.DB.QueryRowContext(
		ctx,
		"SELECT is_admin FROM users WHERE id = $1",
		userID).
		Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrAppNotFound
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.postgres.App"

	var app models.App

	err := s.DB.QueryRowContext(
		ctx,
		"SELECT id, name, secret FROM apps WHERE id = $1",
		appID).
		Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) TokenSecret(ctx context.Context) (string, error) {
	const op = "storage.postgres.TokenSecret"

	var tokenSecret string

	err := s.DB.QueryRowContext(
		ctx,
		"SELECT secret FROM apps",
	).Scan(&tokenSecret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrSecretNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenSecret, nil
}
