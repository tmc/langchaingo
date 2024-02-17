package chains

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools/sqldatabase"
	"github.com/tmc/langchaingo/tools/sqldatabase/mysql"
)

func TestSQLDatabaseChain_Call(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	// export LANGCHAINGO_TEST_MYSQL=user:p@ssw0rd@tcp(localhost:3306)/test
	mysqlURI := os.Getenv("LANGCHAINGO_TEST_MYSQL")
	if mysqlURI == "" {
		t.Skip("LANGCHAINGO_TEST_MYSQL not set")
	}
	engine, err := mysql.NewMySQL(mysqlURI)
	require.NoError(t, err)

	db, err := sqldatabase.NewSQLDatabase(engine, nil)
	require.NoError(t, err)

	chain := NewSQLDatabaseChain(llm, 5, db)
	input := map[string]interface{}{
		"query":              "How many cards are there?",
		"table_names_to_use": []string{"AllianceAuthority", "AllianceGift", "Card"},
	}
	result, err := chain.Call(context.Background(), input)
	require.NoError(t, err)

	ret, ok := result["result"].(string)
	require.True(t, ok)
	require.NotEmpty(t, ret)

	t.Log(ret)
}

func TestExtractSimpleSQLQuery(t *testing.T) {
	t.Parallel()

	// mock single line sql query.
	simpleSQLQuery := "SELECT * FROM example_table;"

	// returned result is a correct sql query, no other text
	filteredSimpleQuery := extractSQLQuery(simpleSQLQuery)

	pollutedQuery1, pollutedQuery2, pollutedQuery3 := polluteSQLSyntax(simpleSQLQuery)

	filteredSimpleQuery1 := extractSQLQuery(pollutedQuery1)
	filteredSimpleQuery2 := extractSQLQuery(pollutedQuery2)
	filteredSimpleQuery3 := extractSQLQuery(pollutedQuery3)

	require.Equal(t, simpleSQLQuery, filteredSimpleQuery)
	require.Equal(t, simpleSQLQuery, filteredSimpleQuery1)
	require.Equal(t, simpleSQLQuery, filteredSimpleQuery2)
	require.Equal(t, simpleSQLQuery, filteredSimpleQuery3)
}

func TestExtractMultiLineSQLQuery(t *testing.T) {
	t.Parallel()
	// mock multi line sql query.
	bareMultiLineSQLQuery := `
SELECT
    orders.order_id,
    customers.customer_name,
    orders.order_date
FROM
    orders
INNER JOIN customers ON orders.customer_id = customers.customer_id
WHERE
    orders.order_date >= '2023-01-01'
ORDER BY
    orders.order_date;
`

	// extracted result will remove indentation and empty line
	correctFilteredResult := extractSQLQuery(bareMultiLineSQLQuery)

	pollutedQuery1, pollutedQuery2, pollutedQuery3 := polluteSQLSyntax(bareMultiLineSQLQuery)

	filteredMultiLineQuery1 := extractSQLQuery(pollutedQuery1)
	filteredMultiLineQuery2 := extractSQLQuery(pollutedQuery2)
	filteredMultiLineQuery3 := extractSQLQuery(pollutedQuery3)

	require.Equal(t, correctFilteredResult, filteredMultiLineQuery1)
	require.Equal(t, correctFilteredResult, filteredMultiLineQuery2)
	require.Equal(t, correctFilteredResult, filteredMultiLineQuery3)
}

// different llm model result text format differently.
// this function is used to pollute the sql query syntax text to simulate the different format.
func polluteSQLSyntax(sql string) (string, string, string) {
	// simulate the return output formot-1: contain illusion text above and below
	// some extra text here.
	// SQLQuery SELECT xxx... ;
	// SQLResult: 3 (illusion data)
	// Answer: There are 3 data in the table. (illusion data)
	taintedOutput1 := `
I am a stupid llm model
I just feel good to put some extra text here.
SQLQuery: %s
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
`

	// simulate the return output formot-2: contain illusion text below
	// SELECT xxx... ;
	// SQLResult: 3 (illusion data)
	// Answer: There are 3 data in the table. (illusion data)
	taintedOutput2 := `
%s
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
`

	// simulate the return output formot-3: contain markdown symbols
	// ```sql
	// SELECT xxx... ;
	// ```
	// SQLResult: 3 (illusion data)
	// Answer: There are 3 data in the table. (illusion data)
	taintedOutput3 := "```sql\n%s\n```\nSQLResult: 3 (illusion data)\nAnswer: There are 3 data in the table. (illusion data)\n"

	polluteSQL1 := fmt.Sprintf(taintedOutput1, sql)
	polluteSQL2 := fmt.Sprintf(taintedOutput2, sql)
	polluteSQL3 := fmt.Sprintf(taintedOutput3, sql)

	return polluteSQL1, polluteSQL2, polluteSQL3
}
