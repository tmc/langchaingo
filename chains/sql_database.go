package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools/sqldatabase"
)

//nolint:lll
const _defaultSQLTemplate = `Given an input question, first create a syntactically correct {{.dialect}} query to run, then look at the results of the query and return the answer. Unless the user specifies in his question a specific number of examples he wishes to obtain, always limit your query to at most {{.top_k}} results. You can order the results by a relevant column to return the most interesting examples in the database.

Never query for all the columns from a specific table, only ask for a the few relevant columns given the question.

Pay attention to use only the column names that you can see in the schema description. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.

Use the following format:

Question: Question here
SQLQuery: SQL Query to run
SQLResult: Result of the SQLQuery
Answer: Final answer here

`

//nolint:lll
const _defaultSQLSuffix = `Only use the following tables:
{{.table_info}}

Question: {{.input}}`

const (
	_sqlChainDefaultInputKeyQuery      = "query"
	_sqlChainDefaultInputKeyTableNames = "table_names_to_use"
	_sqlChainDefaultOutputKey          = "result"
)

// SQLDatabaseChain is a chain used for interacting with SQL Database.
type SQLDatabaseChain struct {
	LLMChain  *LLMChain
	TopK      int
	Database  *sqldatabase.SQLDatabase
	OutputKey string
}

// NewSQLDatabaseChain creates a new SQLDatabaseChain.
// The topK is the max number of results to return.
func NewSQLDatabaseChain(llm llms.Model, topK int, database *sqldatabase.SQLDatabase) *SQLDatabaseChain {
	p := prompts.NewPromptTemplate(_defaultSQLTemplate+_defaultSQLSuffix,
		[]string{"dialect", "top_k", "table_info", "input"})
	c := NewLLMChain(llm, p)
	return &SQLDatabaseChain{
		LLMChain:  c,
		TopK:      topK,
		Database:  database,
		OutputKey: _sqlChainDefaultOutputKey,
	}
}

// Call calls the chain.
// Inputs:
//
//	"query" : key with the query to run.
//	"table_names_to_use" (optionally): a slice string of the only table names
//		to use(others will be ignored).
//
// Outputs
//
//	"result" : with the result of the query.
//
//nolint:all
func (s SQLDatabaseChain) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) {
	query, ok := inputs[_sqlChainDefaultInputKeyQuery].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	var tables []string
	if ts, ok := inputs[_sqlChainDefaultInputKeyTableNames]; ok {
		if tables, ok = ts.([]string); !ok {
			return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
		}
	} else {
		tables = nil
	}

	// Get tables infos
	tableInfos, err := s.Database.TableInfo(ctx, tables)
	if err != nil {
		return nil, err
	}

	const (
		queryPrefixWith = "\nSQLQuery:"  //nolint:gosec
		stopWord        = "\nSQLResult:" //nolint:gosec
	)
	llmInputs := map[string]any{
		"input":      query + queryPrefixWith,
		"top_k":      s.TopK,
		"dialect":    s.Database.Dialect(),
		"table_info": tableInfos,
	}

	// Predict sql query
	opt := append(options, WithStopWords([]string{stopWord})) //nolint:cyclop
	out, err := Predict(ctx, s.LLMChain, llmInputs, opt...)
	if err != nil {
		return nil, err
	}

	sqlQuery := extractSQLQuery(out)

	if sqlQuery == "" {
		return nil, fmt.Errorf("no sql query generated")
	}

	// Execute sql query
	queryResult, err := s.Database.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	// Generate answer
	llmInputs["input"] = query + queryPrefixWith + sqlQuery + stopWord + queryResult
	out, err = Predict(ctx, s.LLMChain, llmInputs, options...)
	if err != nil {
		return nil, err
	}

	// Hack answer string
	strs := strings.Split(strings.Split(out, "\n\n")[0], "Answer:")
	out = strs[0]
	if len(strs) > 1 {
		out = strings.TrimSpace(strs[1])
	}

	return map[string]any{s.OutputKey: out}, nil
}

func (s SQLDatabaseChain) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
}

func (s SQLDatabaseChain) GetInputKeys() []string {
	return []string{_sqlChainDefaultInputKeyQuery}
}

func (s SQLDatabaseChain) GetOutputKeys() []string {
	return []string{s.OutputKey}
}

// sometimes llm model returned result is not only the SQLQuery,
// it also contains some extra text,
// which will cause the entire process to fail.
// this function is used to extract the exact SQLQuery from the result.
// nolint:cyclop
func extractSQLQuery(rawOut string) string {
	outStrings := strings.Split(rawOut, "\n")

	var sqlQuery string
	containSQLQuery := strings.Contains(rawOut, "SQLQuery:")
	findSQLQuery := false

	for _, v := range outStrings {
		line := strings.TrimSpace(v)

		// filter empty line and markdown symbols
		if line == "" || strings.HasPrefix(line, "```") {
			continue
		}

		// stop when we find SQLResult: or Answer:
		if strings.HasPrefix(line, "SQLResult:") || strings.HasPrefix(line, "Answer:") {
			break
		}

		var currentLine string
		switch {
		case containSQLQuery && strings.HasPrefix(line, "SQLQuery:"):
			findSQLQuery = true
			currentLine = strings.TrimPrefix(line, "SQLQuery:")
			if strings.TrimSpace(currentLine) == "" {
				continue
			}
		case containSQLQuery && !findSQLQuery:
			// filter unwanted text above the SQLQuery:
			continue
		default:
			currentLine = line
		}

		sqlQuery += currentLine + "\n"
	}

	return strings.TrimSpace(sqlQuery)
}
