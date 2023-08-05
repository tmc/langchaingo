package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/tmc/langchaingo/tools/sqldatabase"
)

const EngineName = "mysql"

//nolint:gochecknoinits
func init() {
	sqldatabase.RegisterEngine(EngineName, NewMySQL)
}

var _ sqldatabase.Engine = MySQL{}

// MySQL is a MySQL engine.
type MySQL struct {
	db *sql.DB
}

// NewMySQL creates a new MySQL engine.
// The dsn is the data source name.(e.g. root:password@tcp(localhost:3306)/test).
func NewMySQL(dsn string) (sqldatabase.Engine, error) { //nolint:ireturn
	db, err := sql.Open(EngineName, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(32) //nolint:gomnd

	return &MySQL{
		db: db,
	}, nil
}

func (m MySQL) Dialect() string {
	return EngineName
}

func (m MySQL) Query(ctx context.Context, query string, args ...any) ([]string, [][]string, error) {
	rows, err := m.db.QueryContext(ctx, query, args...)
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

func (m MySQL) TableNames(ctx context.Context) ([]string, error) {
	_, result, err := m.Query(ctx, "SHOW TABLES")
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(result))
	for _, row := range result {
		ret = append(ret, row[0])
	}
	return ret, nil
}

func (m MySQL) TableInfo(ctx context.Context, table string) (string, error) {
	_, result, err := m.Query(ctx, fmt.Sprintf("SHOW CREATE TABLE %s", table))
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

func (m MySQL) Close() error {
	return m.db.Close()
}
