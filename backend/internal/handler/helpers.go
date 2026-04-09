package handler

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// pgTimestamp converts a time.Time to pgtype.Timestamptz for use in sqlc queries.
func pgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
