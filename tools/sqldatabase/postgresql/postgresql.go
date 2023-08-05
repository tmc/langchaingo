package postgresql

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // postgresql driver
	"github.com/tmc/langchaingo/tools/sqldatabase"
)

const EngineName = "pgx"

// init registers the PostgreSQL engine with the sqldatabase package.
// It is automatically called during package initialization.
//
//nolint:gochecknoinits
func init() {
	sqldatabase.RegisterEngine(EngineName, NewPostgreSQL)
}

var _ sqldatabase.Engine = PostgreSQL{}

// PostgreSQL represents the PostgreSQL engine.
type PostgreSQL struct {
	db *sql.DB
}

// NewPostgreSQL creates a new PostgreSQL engine instance.
// The dsn parameter is the data source name
// (e.g. postgres://db_user:mysecretpassword@localhost:5438/test?sslmode=disable).
// It returns a sqldatabase.Engine and an error, if any.
func NewPostgreSQL(dsn string) (sqldatabase.Engine, error) { //nolint:ireturn
	db, err := sql.Open(EngineName, dsn)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		db: db,
	}, nil
}

// Dialect returns the dialect of the PostgreSQL engine.
func (p PostgreSQL) Dialect() string {
	return EngineName
}

// Query executes a query on the PostgreSQL engine.
// It takes a context.Context, a query string, and optional query arguments.
// It returns the column names, query results as a 2D slice of strings, and an error, if any.
func (p PostgreSQL) Query(ctx context.Context, query string, args ...any) ([]string, [][]string, error) {
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	results := make([][]string, 0)
	for rows.Next() {
		row := make([]string, len(cols))
		rowNullable := make([]sql.NullString, len(cols))
		rowPtrs := make([]interface{}, len(cols))
		for i := range row {
			rowPtrs[i] = &rowNullable[i]
		}
		err = rows.Scan(rowPtrs...)
		if err != nil {
			return nil, nil, err
		}
		for _, v := range rowNullable {
			row = append(row, v.String)
		}
		results = append(results, row)
	}
	return cols, results, nil
}

// TableNames returns the names of all tables in the PostgreSQL database.
// It takes a context.Context.
// It returns a slice of table names and an error, if any.
func (p PostgreSQL) TableNames(ctx context.Context) ([]string, error) {
	_, result, err := p.Query(ctx,
		`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(result))
	for _, row := range result {
		ret = append(ret, row[0])
	}
	return ret, nil
}

// TableInfo returns information about a specific table in the PostgreSQL database.
// It takes a context.Context and the name of the table.
// It returns the table name and an error, if any.
func (p PostgreSQL) TableInfo(ctx context.Context, table string) (string, error) {
	_, result, err := p.Query(ctx, `SELECT 
		table_name, 
		column_name, 
		data_type 
	 FROM 
		information_schema.columns
	 WHERE 
		table_name = $1`, table)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", sqldatabase.ErrTableNotFound
	}
	if len(result[0]) < 2 { //nolint:gomnd
		return "", sqldatabase.ErrInvalidResult
	}

	return result[0][1], nil //nolint:gomnd
}

// Close closes the connection to the PostgreSQL database.
// It returns an error, if any.
func (p PostgreSQL) Close() error {
	return p.db.Close()
}
