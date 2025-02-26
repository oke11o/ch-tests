package migrator

import (
	"context"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/oke11o/ch-tests/internal/config"
)

type Migrator struct {
	cfg config.ClickHouseConfig
	m   *migrate.Migrate
}

func NewMigrator(_ context.Context, cfg config.ClickHouseConfig, migPath string) (*Migrator, error) {
	sourceURL := "file://" + migPath
	if len(cfg.Host) < 1 {
		return nil, fmt.Errorf("migrator host must be provided")
	}
	databaseURL := cfg.Dsn()

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{
		cfg: cfg,
		m:   m,
	}, nil
}

func (m *Migrator) Run() error {
	if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
