package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tmc/langchaingo/tools/sql_database"
)

func init() {
	sql_database.RegisterEngine("mysql", NewMySQL)
}

var _ sql_database.Engine = MySQL{}

// MySQL is a MySQL engine.
type MySQL struct {
	db *sql.DB
}

// NewMySQL creates a new MySQL engine.
// The dsn is the data source name.(e.g. root:password@tcp(localhost:3306)/test)
func NewMySQL(dsn string) (sql_database.Engine, error) { //nolint:ireturn
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(32)

	return &MySQL{
		db: db,
	}, nil
}

func (m MySQL) Dialect() string {
	return "mysql"
}

func (m MySQL) Query(ctx context.Context, query string) (cols []string, results [][]string, err error) {
	rows, err := m.db.QueryContext(ctx, query)
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
	_, result, err := m.Query(ctx, "SHOW CREATE TABLE "+table)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", fmt.Errorf("table %s not found", table)
	}
	if len(result[0]) < 2 {
		return "", fmt.Errorf("invalid result %v", result)
	}

	return result[0][1], nil
}
