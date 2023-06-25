package mysql_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/tmc/langchaingo/tools/sql_database"
	_ "github.com/tmc/langchaingo/tools/sql_database/mysql"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.Background()
)

func Test(t *testing.T) {
	err := godotenv.Overload()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_DATABASE"))
	db, err := sql_database.NewSQLDatabaseWithDSN("mysql", dsn, nil)
	require.NoError(t, err)

	tbs := db.TableNames()
	require.Greater(t, len(tbs), 0)

	desc, err := db.TableInfo(ctx, tbs)
	require.NoError(t, err)

	t.Log(desc)
}
