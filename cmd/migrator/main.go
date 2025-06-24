package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"strings"
)

func main() {
	var dsn, migrationsPath, migrationsTable string

	flag.StringVar(&dsn, "dsn", "", "postgres DSN")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations")
	flag.Parse()

	if dsn == "" {
		panic("dsn is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	var dbURL string
	if strings.Contains(dsn, "?") {
		dbURL = dsn + "&x-migrations-table=" + migrationsTable
	} else {
		dbURL = dsn + "?x-migrations-table=" + migrationsTable
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		dbURL,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied successfully")
}
