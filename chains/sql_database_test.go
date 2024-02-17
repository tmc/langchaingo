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

// mock single line sql query
var simpleSqlQuery = "SELECT * FROM example_table;"

// mock multi line sql query
var bareMultiLineSqlQuery = `
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
var multiLineQueryExtractedResult = `SELECT
orders.order_id,
customers.customer_name,
orders.order_date
FROM
orders
INNER JOIN customers ON orders.customer_id = customers.customer_id
WHERE
orders.order_date >= '2023-01-01'
ORDER BY
orders.order_date;`

/*
simulate the return output formot-1: contain illusion text
SQLQuery SELECT xxx... ;
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
*/
var taintedOutput1 = `
SQLQuery: %s
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
`

/*
simulate the return output formot-2: contain illusion text
SELECT xxx... ;
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
*/
var taintedOutput2 = `
%s
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
`

/*
simulate the return output formot-3: contain markdown symbols
```sql
SELECT xxx... ;
```
SQLResult: 3 (illusion data)
Answer: There are 3 data in the table. (illusion data)
*/
var taintedOutput3 = "```sql\n%s\n```\nSQLResult: 3 (illusion data)\nAnswer: There are 3 data in the table. (illusion data)\n"

func TestExtractSimpleSqlQuery(t *testing.T) {

	t.Parallel()

	filteredSimpleQuery := extractSqlQuery(simpleSqlQuery)

	simpleSqlQueryWithUnwantedText1 := fmt.Sprintf(taintedOutput1, simpleSqlQuery)
	filteredSimpleQuery1 := extractSqlQuery(simpleSqlQueryWithUnwantedText1)

	simpleSqlQueryWithUnwantedText2 := fmt.Sprintf(taintedOutput2, simpleSqlQuery)
	filteredSimpleQuery2 := extractSqlQuery(simpleSqlQueryWithUnwantedText2)

	simpleSqlQueryWithUnwantedText3 := fmt.Sprintf(taintedOutput3, simpleSqlQuery)
	filteredSimpleQuery3 := extractSqlQuery(simpleSqlQueryWithUnwantedText3)

	require.Equal(t, simpleSqlQuery, filteredSimpleQuery)
	require.Equal(t, simpleSqlQuery, filteredSimpleQuery1)
	require.Equal(t, simpleSqlQuery, filteredSimpleQuery2)
	require.Equal(t, simpleSqlQuery, filteredSimpleQuery3)
}

func TestExtractMultiLineSqlQuery(t *testing.T) {
	t.Parallel()

	filteredMultiLineQuery := extractSqlQuery(bareMultiLineSqlQuery)

	multiLineSqlQueryWithUnwantedText1 := fmt.Sprintf(taintedOutput1, bareMultiLineSqlQuery)
	filteredMultiLineQuery1 := extractSqlQuery(multiLineSqlQueryWithUnwantedText1)

	multiLineSqlQueryWithUnwantedText2 := fmt.Sprintf(taintedOutput2, bareMultiLineSqlQuery)
	filteredMultiLineQuery2 := extractSqlQuery(multiLineSqlQueryWithUnwantedText2)

	multiLineSqlQueryWithUnwantedText3 := fmt.Sprintf(taintedOutput2, bareMultiLineSqlQuery)
	filteredMultiLineQuery3 := extractSqlQuery(multiLineSqlQueryWithUnwantedText3)

	require.Equal(t, multiLineQueryExtractedResult, filteredMultiLineQuery)
	require.Equal(t, multiLineQueryExtractedResult, filteredMultiLineQuery1)
	require.Equal(t, multiLineQueryExtractedResult, filteredMultiLineQuery2)
	require.Equal(t, multiLineQueryExtractedResult, filteredMultiLineQuery3)
}
