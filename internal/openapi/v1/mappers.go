package v1

import (
	"github.com/dariomba/screen-go/internal/openapi"
	"github.com/dariomba/screen-go/internal/postgres"
	"github.com/jackc/pgx/v5/pgtype"
)

// toCreateJobParams maps OpenAPI request to sqlc params
func toCreateJobParams(id string, req *openapi.CreateJobJSONRequestBody) postgres.CreateJobParams {
	return postgres.CreateJobParams{
		ID:       id,
		Url:      req.URL,
		Format:   toPgNullJobFormat(req.Format),
		Width:    toPgInt4(req.Width),
		Height:   toPgInt4(req.Height),
		FullPage: toPgBool(req.FullPage),
	}
}

// toPgInt4 converts *int to pgtype.Int4 (nullable)
func toPgInt4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

// toPgBool converts *bool to pgtype.Bool (nullable)
func toPgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func toPgNullJobFormat(f *openapi.CreateJobRequestFormat) postgres.NullJobFormat {
	if f == nil {
		return postgres.NullJobFormat{Valid: false}
	}

	return postgres.NullJobFormat{
		JobFormat: postgres.JobFormat(*f),
		Valid:     true,
	}
}
