package service

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/oke11o/ch-tests/internal/config"
	"github.com/oke11o/ch-tests/internal/migrator"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
)

type ClickHouseTestSuite struct {
	suite.Suite
	pool     *dockertest.Pool
	resource *dockertest.Resource
	conn     *sql.DB
}

func (s *ClickHouseTestSuite) SetupSuite() {
	ctx := context.Background()
	cfg, err := config.NewConfig(ctx, "TEST")

	s.Require().NoError(err)
	cfg.CH.Db = "testdb"

	_, currentFile, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(currentFile), "../../migrations/clickhouse")

	pool, err := dockertest.NewPool("")
	s.Require().NoError(err, "Failed to create Docker pool")
	s.pool = pool

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "clickhouse/clickhouse-server",
		Tag:        "latest",
		Env: []string{
			"CLICKHOUSE_USER=" + cfg.CH.User,
			"CLICKHOUSE_PASSWORD=" + cfg.CH.Password,
			"CLICKHOUSE_DB=" + cfg.CH.Db,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	s.Require().NoError(err, "Failed to start ClickHouse container")
	s.resource = resource
	host := resource.GetBoundIP("9000/tcp")
	port := resource.GetPort("9000/tcp")
	cfg.CH.Host = []string{fmt.Sprintf("%s:%s", host, port)}
	s.T().Log("ClickHouse DSN ", cfg.CH.Dsn())
	s.T().Log("resource.GetBoundIP - 8123 ", resource.GetBoundIP("8123/tcp"))

	// Подключаемся к дефолтной базе для создания test_db_123
	tmp := fmt.Sprintf("clickhouse://%s:%s/%s?username=%s&password=%s", host, port, cfg.CH.Db, cfg.CH.User, cfg.CH.Password)
	dsn := cfg.CH.Dsn()
	s.Require().Equal(tmp, dsn)

	// Ждем готовности и создаем базу
	err = pool.Retry(func() error {
		s.conn, err = sql.Open("clickhouse", dsn)
		if err != nil {
			return err
		}
		return s.conn.Ping()
	})
	s.Require().NoError(err, "Failed to connect to ClickHouse")

	// Migrations
	mig, err := migrator.NewMigrator(ctx, cfg.CH, migrationsPath)
	s.Require().NoError(err)
	err = mig.Run()
	s.Require().NoError(err)
}

func (s *ClickHouseTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.resource != nil {
		s.Require().NoError(s.pool.Purge(s.resource), "Failed to purge container")
	}
	s.T().Log("TESTS DONE")
}

func (s *ClickHouseTestSuite) TestFirst() {
	var count int
	err := s.conn.QueryRow("SELECT COUNT(*) FROM user_logs").Scan(&count)
	s.Require().NoError(err, "Failed to count rows in user_logs table")
	s.Require().Equal(0, count, "user_logs table should be empty")
}

func TestClickHouseSuite(t *testing.T) {
	suite.Run(t, new(ClickHouseTestSuite))
}
