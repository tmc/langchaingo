package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools/sql_database"
)

//nolint:lll
const _defaultSqlTemplate = `Given an input question, first create a syntactically correct {{.dialect}} query to run, then look at the results of the query and return the answer. Unless the user specifies in his question a specific number of examples he wishes to obtain, always limit your query to at most {{.top_k}} results. You can order the results by a relevant column to return the most interesting examples in the database.

Never query for all the columns from a specific table, only ask for a the few relevant columns given the question.

Pay attention to use only the column names that you can see in the schema description. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.

Use the following format:

Question: Question here
SQLQuery: SQL Query to run
SQLResult: Result of the SQLQuery
Answer: Final answer here

`

//nolint:lll
const _defaultMySqlTemplate = `You are a MySQL expert. Given an input question, first create a syntactically correct MySQL query to run, then look at the results of the query and return the answer to the input question.
Unless the user specifies in the question a specific number of examples to obtain, query for at most {{.top_k}} results using the LIMIT clause as per MySQL. You can order the results to return the most informative data in the database.
Never query for all columns from a table. You must query only the columns that are needed to answer the question. Wrap each column name in backticks (` + "`" + `) to denote them as delimited identifiers.
Pay attention to use only the column names you can see in the tables below. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.
Pay attention to use CURDATE() function to get the current date, if the question involves "today".

Use the following format:

Question: Question here
SQLQuery: SQL Query to run
SQLResult: Result of the SQLQuery
Answer: Final answer here

`

//nolint:lll
const _defaultSqlSuffix = `Only use the following tables:
{{.table_info}}

Question: {{.input}}`

// SQLDatabaseChain is a chain used for interacting with SQL Database.
type SQLDatabaseChain struct {
	LLMChain *LLMChain
	TopK     int
	Database *sql_database.SQLDatabase
}

// NewSQLDatabaseChain creates a new SQLDatabaseChain.
// The topK is the max number of results to return.
func NewSQLDatabaseChain(llm llms.LLM, topK int, database *sql_database.SQLDatabase) *SQLDatabaseChain {
	p := prompts.NewPromptTemplate(_defaultSqlTemplate+_defaultSqlSuffix, []string{"dialect", "top_k", "table_info", "input"})
	c := NewLLMChain(llm, p)
	return &SQLDatabaseChain{
		LLMChain: c,
		TopK:     topK,
		Database: database,
	}
}

// Call calls the chain.
// Inputs:
//
//	"query" : key with the query to run.
//	"table_names_to_use" (optionally): the only table names(others will be ignored) to use with a comma separated list of.
//
// Outputs
//
//	"result" : with the result of the query.
func (s SQLDatabaseChain) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) {
	query, ok := inputs["query"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	tables := make([]string, 0, len(s.Database.TableNames()))
	if ts, ok := inputs["table_names_to_use"]; ok {
		strs := strings.Split(ts.(string), ",")
		for _, s := range strs {
			s = strings.TrimSpace(s)
			if len(s) > 0 {
				tables = append(tables, s)
			}
		}

	}

	// Get tables infos
	tableInfos, err := s.Database.TableInfo(ctx, tables)
	if err != nil {
		return nil, err
	}

	const (
		queryPrefixWith = "\nSQLQuery:"
		stopWord        = "\nSQLResult:"
	)
	llmInputs := map[string]any{
		"input":      query + queryPrefixWith,
		"top_k":      s.TopK,
		"dialect":    s.Database.Dialect(),
		"table_info": tableInfos,
	}

	// Predict sql query
	sqlQuery, err := s.LLMChain.Predict(ctx, llmInputs, WithStopWords([]string{stopWord}))
	if err != nil {
		return nil, err
	}
	sqlQuery = strings.TrimSpace(sqlQuery)

	// Execute sql query
	queryResult, err := s.Database.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	// Generate answer
	llmInputs["input"] = query + queryPrefixWith + sqlQuery + stopWord + queryResult
	result, err := s.LLMChain.Predict(ctx, llmInputs)
	if err != nil {
		return nil, err
	}
	// Hack answer string
	strs := strings.Split(strings.Split(result, "\n")[0], "Answer:")
	result = strs[0]
	if len(strs) > 1 {
		result = strings.TrimSpace(strs[1])
	}

	return map[string]any{"result": result}, nil
}

func (s SQLDatabaseChain) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
}

func (s SQLDatabaseChain) GetInputKeys() []string {
	return []string{"query"}
}

func (s SQLDatabaseChain) GetOutputKeys() []string {
	return []string{"result"}
}
