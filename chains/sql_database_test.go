package chains

import (
	"context"
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

func TestExtractSQLQuery(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputStr string
		expected string
	}{
		{
			inputStr: "SELECT * FROM example_table;",
			expected: "SELECT * FROM example_table;",
		},
		{
			inputStr: `
			I am a clumsy llm model
			I just feel good to put some extra text here.

			SQLQuery: SELECT * FROM example_table;
			SQLResult: 3 (this is not a real data)
			Answer: There are 3 data in the table. (this is not a real data)`,
			expected: "SELECT * FROM example_table;",
		},
		{
			inputStr: `
			SELECT * FROM example_table;
			SQLResult: 3 (this is not a real data)
			Answer: There are 3 data in the table. (this is not a real data)`,
			expected: "SELECT * FROM example_table;",
		},
		{
			inputStr: "```sql" + `
			SELECT * FROM example_table;
			` + "```" + `
			SQLResult: 3 (this is not a real data)
			Answer: There are 3 data in the table. (this is not a real data)`,
			expected: "SELECT * FROM example_table;",
		},
		{ // multi-line sql query with markdown symbols and redundant text above and below
			inputStr: `
			I am also a clumsy llm model, I don't fully understand the prompt
			And accidentally put some extra text here.

			SQLQuery: 
			` + "```sql\n" + `
			SELECT
				order_id,
				customer_name,
				order_date
			FROM orders;
			` + "```" + `
			SQLResult: xxx (this is not a real data)
			Answer: some illusion answer. (this is not a real data)`,
			expected: `SELECT
order_id,
customer_name,
order_date
FROM orders;`,
		},
		{ // slightly complexed multi-line query, no extra text before but only with redundant text after
			inputStr: `SELECT
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
			SQLResult: xxx (this is not a real data)
			Answer: some illusion answer. (this is not a real data)`,
			expected: `SELECT
orders.order_id,
customers.customer_name,
orders.order_date
FROM
orders
INNER JOIN customers ON orders.customer_id = customers.customer_id
WHERE
orders.order_date >= '2023-01-01'
ORDER BY
orders.order_date;`,
		},
	}

	for _, tc := range cases {
		filterQuerySyntax := extractSQLQuery(tc.inputStr)
		require.Equal(t, tc.expected, filterQuerySyntax)
	}
}
