package sqldatabase

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// EngineFunc is the function that returns the database engine.
type EngineFunc func(string) (Engine, error)

//nolint:gochecknoglobals
var engines = make(map[string]EngineFunc)

func RegisterEngine(name string, engineFunc EngineFunc) {
	engines[name] = engineFunc
}

// Engine is the interface that wraps the database.
type Engine interface {
	// Dialect returns the dialect(e.g. mysql, sqlite, postgre) of the database.
	Dialect() string

	// Query executes the query and returns the columns and results.
	Query(ctx context.Context, query string, args ...any) (cols []string, results [][]string, err error)

	// TableNames returns all the table names of the database.
	TableNames(ctx context.Context) ([]string, error)

	// TableInfo returns the table information of the database.
	// Typically, it returns the CREATE TABLE statement.
	TableInfo(ctx context.Context, tables string) (string, error)

	// Close closes the database.
	Close() error
}

var (
	ErrUnknownDialect = fmt.Errorf("unknown dialect")

	ErrTableNotFound = fmt.Errorf("table not found")
	ErrInvalidResult = fmt.Errorf("invalid result")
)

// SQLDatabase sql wrapper.
type SQLDatabase struct {
	Engine           Engine // The database engine.
	SampleRowsNumber int    // The number of sample rows to show. 0 means no sample rows.
	allTables        []string
}

// NewSQLDatabase creates a new SQLDatabase.
func NewSQLDatabase(engine Engine, ignoreTables map[string]struct{}) (*SQLDatabase, error) {
	sd := &SQLDatabase{
		Engine:           engine,
		SampleRowsNumber: 3, //nolint:gomnd
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:gomnd
	defer cancel()
	tbs, err := engine.TableNames(ctx)
	if err != nil {
		return nil, err
	}
	for _, tb := range tbs {
		if _, ok := ignoreTables[tb]; ok {
			continue
		}
		sd.allTables = append(sd.allTables, tb)
	}

	return sd, nil
}

// NewSQLDatabaseWithDSN creates a new SQLDatabase with the data source name.
func NewSQLDatabaseWithDSN(dialect, dsn string, ignoreTables map[string]struct{}) (*SQLDatabase, error) {
	engineFunc, ok := engines[dialect]
	if !ok {
		return nil, ErrUnknownDialect
	}
	engine, err := engineFunc(dsn)
	if err != nil {
		return nil, err
	}
	return NewSQLDatabase(engine, ignoreTables)
}

// Dialect returns the dialect(e.g. mysql, sqlite, postgre) of the database.
func (sd *SQLDatabase) Dialect() string {
	return sd.Engine.Dialect()
}

// TableNames returns all the table names of the database.
func (sd *SQLDatabase) TableNames() []string {
	return sd.allTables
}

// TableInfo returns the table information string of the database.
// If tables is empty, it will return all the tables, otherwise it will return the given tables.
func (sd *SQLDatabase) TableInfo(ctx context.Context, tables []string) (string, error) {
	if len(tables) == 0 {
		tables = sd.allTables
	}
	str := ""
	for _, tb := range tables {
		// Get table info
		info, err := sd.Engine.TableInfo(ctx, tb)
		if err != nil {
			return "", err
		}
		str += info + "\n\n"

		// Get sample rows
		if sd.SampleRowsNumber > 0 {
			sampleRows, err := sd.sampleRows(ctx, tb, sd.SampleRowsNumber)
			if err != nil {
				return "", err
			}
			str += "/*\n" + sampleRows + "*/ \n\n"
		}
	}

	return str, nil
}

// Query executes the query and returns the string that contains columns and results.
func (sd *SQLDatabase) Query(ctx context.Context, query string) (string, error) {
	cols, results, err := sd.Engine.Query(ctx, query)
	if err != nil {
		return "", err
	}

	str := strings.Join(cols, "\t") + "\n"
	for _, row := range results {
		str += strings.Join(row, "\t") + "\n"
	}
	return str, nil
}

// Close closes the database.
func (sd *SQLDatabase) Close() error {
	return sd.Engine.Close()
}

func (sd *SQLDatabase) sampleRows(ctx context.Context, table string, rows int) (string, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", table, rows)
	result, err := sd.Query(ctx, query)
	if err != nil {
		return "", err
	}
	ret := fmt.Sprintf("%d rows from %s table:\n", rows, table)
	ret += result
	return ret, nil
}
