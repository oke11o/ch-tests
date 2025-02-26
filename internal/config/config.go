package config

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CH            ClickHouseConfig `envconfig:"ch"`
	MigrationPath string           `envconfig:"migrationpath" default:"migrations/clickhouse"`
}

type ClickHouseConfig struct {
	Host     []string `default:"localhost:9000"`
	User     string   `default:"user"`
	Password string   `default:"password"`
	Db       string   `default:"mydb"`
}

func (c *ClickHouseConfig) Dsn() string {
	host := c.Host[0]
	databaseURL := fmt.Sprintf("clickhouse://%s/%s?username=%s&password=%s",
		host, c.Db, c.User, c.Password)

	return databaseURL
}

func NewConfig(ctx context.Context, prefix string) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(prefix, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
