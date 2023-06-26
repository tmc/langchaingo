package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
	"github.com/tmc/langchaingo/tools/sqldatabase"
)

const name = "sqlite3"

func init() {
	sqldatabase.RegisterEngine(name, NewSQLite)
}

var _ sqldatabase.Engine = SQLite{}

// SQLite is a SQLite engine.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite engine.
// The dsn is the data source name.(e.g. file:locked.sqlite?cache=shared)
func NewSQLite(dsn string) (sqldatabase.Engine, error) { //nolint:ireturn
	db, err := sql.Open(name, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	return &SQLite{
		db: db,
	}, nil
}

func (m SQLite) Dialect() string {
	return name
}

func (m SQLite) Query(ctx context.Context, query string, args ...any) (cols []string, results [][]string, err error) {
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	cols, err = rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		row := make([]string, len(cols))
		rowPtrs := make([]interface{}, len(cols))
		for i := range row {
			rowPtrs[i] = &row[i]
		}
		err = rows.Scan(rowPtrs...)
		if err != nil {
			return nil, nil, err
		}
		results = append(results, row)
	}
	return
}

func (m SQLite) TableNames(ctx context.Context) ([]string, error) {
	_, result, err := m.Query(ctx, "SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(result))
	for _, row := range result {
		ret = append(ret, row[0])
	}
	return ret, nil
}

func (m SQLite) TableInfo(ctx context.Context, table string) (string, error) {
	_, result, err := m.Query(ctx, "SELECT sql FROM sqlite_master WHERE type='table' AND name=?;", table)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", fmt.Errorf("table %s not found", table)
	}
	if len(result[0]) < 1 {
		return "", fmt.Errorf("invalid result %v", result)
	}

	return result[0][0], nil
}

func (m SQLite) Close() error {
	return m.db.Close()
}
