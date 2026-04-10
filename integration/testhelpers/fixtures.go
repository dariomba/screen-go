//go:build integration

package testhelpers

import (
	"context"
	"database/sql"
)

func TruncateTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx,
		"TRUNCATE TABLE screenshots, jobs RESTART IDENTITY CASCADE",
	)
	return err
}
