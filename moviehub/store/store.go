package store

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")
)

type Store struct {
	Pool *pgxpool.Pool
}

// รหัสข้อผิดพลาดของ PostgreSQL
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func pgCode(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.Code
	}
	return ""
}

func isUnique(err error) bool { return pgCode(err) == pgUniqueViolation }
func isFK(err error) bool     { return pgCode(err) == pgForeignKeyViolation }
