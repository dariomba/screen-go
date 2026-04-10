//go:build integration

package testhelpers

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Containers struct {
	PgConnStr string
	pg        *postgres.PostgresContainer
}

func StartContainers(ctx context.Context) (*Containers, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("screengodb_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &Containers{
		PgConnStr: connStr,
		pg:        pgContainer,
	}, nil
}

func (c *Containers) Terminate(ctx context.Context) {
	c.pg.Terminate(ctx)
}
