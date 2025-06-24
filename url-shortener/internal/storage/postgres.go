package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/lostmyescape/url-shortener/internal/config"
	"log"
)

type Storage struct {
	DB *sql.DB
}

// NewStorage соберет и вернет объект storage
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
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}
	log.Println("Подключение к postgres установлено!")

	// Выполняем SQL-запрос на создание таблицы
	createTable := `
    CREATE TABLE IF NOT EXISTS url (
        id SERIAL PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL UNIQUE
    );
    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
    `
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании таблицы url: %w", err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveUrl"

	var id int64
	query := "INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id"

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
		return "", fmt.Errorf("url not found: %w", err)
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
