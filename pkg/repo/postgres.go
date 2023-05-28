package repo

import "database/sql"

type SqlQueryable interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...any) *sql.Row
}
