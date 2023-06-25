package chains

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools/sql_database"
	"github.com/tmc/langchaingo/tools/sql_database/mysql"
)

func TestSQLDatabaseChain_Call(t *testing.T) {
	godotenv.Overload()

	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	mysqlURI := os.Getenv("LANGCHAINGO_TEST_MYSQLURI")
	if mysqlURI == "" {
		t.Skip("LANGCHAINGO_TEST_MYSQL_URI not set")
	}
	engine, err := mysql.NewMySQL(mysqlURI)
	require.NoError(t, err)

	db, err := sql_database.NewSQLDatabase(engine, nil)
	require.NoError(t, err)

	chain := NewSQLDatabaseChain(llm, 5, db)
	result, err := chain.Call(context.Background(), map[string]interface{}{"query": "总共有几张卡牌", "table_names_to_use": "Card"})
	require.NoError(t, err)

	ret := result["result"].(string)
	require.Greater(t, len(ret), 0)

	t.Log(ret)
}
