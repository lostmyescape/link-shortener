package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/lostmyescape/link-shortener/common/logger/sl"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/config"
)

type Storage struct {
	DB *sql.DB
}

// NewStorage соберет и вернет объект storage
func NewStorage(ctx context.Context, cfg *config.Config, log *slog.Logger) *Storage {
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			panic("timeout waiting for postgresql")
		case <-ticker.C:
			conn, err := connect(cfg)
			if err == nil {
				log.Info("postgresql connected successfully")
				return conn
			}
			log.Error("postgresql not ready, retrying...", sl.Err(err))
		}
	}
}

func connect(cfg *config.Config) (*Storage, error) {
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
		return nil, fmt.Errorf("postgresql connection error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgresql ping failed: %w", err)
	}

	log.Println("successful database connection")

	return &Storage{DB: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveUrl"

	var id int64
	query := `INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id`

	err := s.DB.QueryRow(query, urlToSave, alias).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				if pqErr.Constraint == "url_url_key" {
					return 0, ErrURLExists
				}
				if pqErr.Constraint == "url_alias_key" {
					return 0, ErrAliasExists
				}
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUrl(alias string) (string, error) {
	const op = "storage.postgres.GetUrl"

	var urlString string

	err := s.DB.QueryRow(`SELECT url FROM url WHERE alias = $1`, alias).Scan(&urlString)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return urlString, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	result, err := s.DB.Exec(`DELETE FROM url WHERE alias = $1`, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return ErrAliasNotFound
	}

	return nil
}
