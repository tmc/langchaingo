package sqlite3_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/tools/sql_database"
	_ "github.com/tmc/langchaingo/tools/sql_database/sqlite3"

	"github.com/stretchr/testify/require"
)

var (
	ctx = context.Background()
)

func Test(t *testing.T) {
	const dsn = `test.sqlite`
	os.Remove(dsn)
	defer os.Remove(dsn)

	// Create some example data
	tmpDB, err := sql.Open("sqlite3", dsn+"?cache=shared")
	require.NoError(t, err)

	_, err = tmpDB.Exec("CREATE TABLE `Activity` (\n  `Id` int,\n  `StringId` text,\n  `Note` text,\n  `TimeType` text,\n  `DayOfWeek` text,\n  `Year` text,\n  `Month` text,\n  `Day` text,\n  `Hour` text,\n  `Minute` text,\n  `Second` text,\n  `Duration` int\n) ")
	require.NoError(t, err)
	_, err = tmpDB.Exec("CREATE TABLE `Activity1` (\n  `Id` int,\n  `StringId` text,\n  `Note` text,\n  `TimeType` text,\n  `DayOfWeek` text,\n  `Year` text,\n  `Month` text,\n  `Day` text,\n  `Hour` text,\n  `Minute` text,\n  `Second` text,\n  `Duration` int\n)  ")
	require.NoError(t, err)
	_, err = tmpDB.Exec("CREATE TABLE `Activity2` (\n  `Id` int,\n  `StringId` text,\n  `Note` text,\n  `TimeType` text,\n  `DayOfWeek` text,\n  `Year` text,\n  `Month` text,\n  `Day` text,\n  `Hour` text,\n  `Minute` text,\n  `Second` text,\n  `Duration` int\n)  ")
	require.NoError(t, err)
	tmpDB.Close()

	db, err := sql_database.NewSQLDatabaseWithDSN("sqlite3", dsn, nil)
	require.NoError(t, err)
	defer db.Close()

	tbs := db.TableNames()
	require.Equal(t, len(tbs), 3)

	desc, err := db.TableInfo(ctx, tbs)
	require.NoError(t, err)

	desc = strings.TrimSpace(desc)
	require.True(t, 0 == strings.Index(desc, "CREATE TABLE"))
	require.True(t, strings.Contains(desc, "Activity"))
	require.True(t, strings.Contains(desc, "Activity1"))
	require.True(t, strings.Contains(desc, "Activity2"))
}
