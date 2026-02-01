package httpapi

import "github.com/jackc/pgx/v5/pgtype"

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return u.String()
}

func ptr[T any](v T) *T { return &v }
