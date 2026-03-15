package postgres

import (
	"time"

	"github.com/dariomba/screen-go/internal/adapters/postgres/sqlc"
	"github.com/dariomba/screen-go/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

// toPgInt4 converts int to pgtype.Int4 (nullable)
func toPgInt4(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// toPgBool converts bool to pgtype.Bool (nullable)
func toPgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func toPgNullJobFormat(f domain.JobFormat) sqlc.NullJobFormat {
	return sqlc.NullJobFormat{
		JobFormat: sqlc.JobFormat(f),
		Valid:     true,
	}
}

func fromPgNullText(s pgtype.Text) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func fromPgNullTime(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
