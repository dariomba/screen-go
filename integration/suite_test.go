//go:build integration

package integration

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"

	"github.com/dariomba/screen-go/integration/testhelpers"
	"github.com/dariomba/screen-go/internal/app"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

type BaseSuite struct {
	suite.Suite
	ctx        context.Context
	containers *testhelpers.Containers
	DB         *sql.DB
}

func (s *BaseSuite) SetupSuite() {
	s.ctx = context.Background()

	containers, err := testhelpers.StartContainers(s.ctx)
	s.Require().NoError(err, "failed to start test containers")
	s.containers = containers

	db, err := sql.Open("postgres", containers.PgConnStr)
	s.Require().NoError(err)
	s.Require().NoError(db.PingContext(s.ctx))
	s.DB = db

	s.runMigrations()
}

func (s *BaseSuite) TearDownSuite() {
	s.DB.Close()
	s.containers.Terminate(s.ctx)
}

func (s *BaseSuite) TearDownTest() {
	err := testhelpers.TruncateTables(s.ctx, s.DB)
	s.Require().NoError(err, "failed to truncate tables between tests")
}

func (s *BaseSuite) runMigrations() {
	m, err := migrate.New(
		"file://../tools/migrate",
		s.containers.PgConnStr,
	)
	s.Require().NoError(err, "failed to create migrate instance")

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.Require().NoError(err, "failed to run migrations")
	}
}

type APIBaseSuite struct {
	BaseSuite
	Server *httptest.Server
}

func (s *APIBaseSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	ctr := app.NewContainer()
	ctr.DBConnStr = s.containers.PgConnStr

	s.Server = httptest.NewServer(ctr.HTTPServer().Handler)
}

func (s *APIBaseSuite) TearDownSuite() {
	s.Server.Close()
	s.BaseSuite.TearDownSuite()
}
