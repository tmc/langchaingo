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

// SQLDatabaseChain is a chain used for interacting with SQL Database.
type SQLDatabaseChain struct {
	LLMChain *LLMChain
	TopK     int
	Database *sqldatabase.SQLDatabase
}

// NewSQLDatabaseChain creates a new SQLDatabaseChain.
// The topK is the max number of results to return.
func NewSQLDatabaseChain(llm llms.LanguageModel, topK int, database *sqldatabase.SQLDatabase) *SQLDatabaseChain {
	p := prompts.NewPromptTemplate(_defaultSQLTemplate+_defaultSQLSuffix,
		[]string{"dialect", "top_k", "table_info", "input"})
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
//	"table_names_to_use" (optionally): the only table names(others will be ignored)
//		to use with a comma separated list of.
//
// Outputs
//
//	"result" : with the result of the query.
func (s SQLDatabaseChain) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint: lll
	query, ok := inputs["query"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	var tables []string
	if ts, ok := inputs["table_names_to_use"]; ok {
		tables = make([]string, 0, len(s.Database.TableNames()))
		strs := strings.Split(ts.(string), ",") //nolint:forcetypeassert
		for _, s := range strs {
			s = strings.TrimSpace(s)
			if len(s) > 0 {
				tables = append(tables, s)
			}
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
	opt := append(options, WithStopWords([]string{stopWord}))
	out, err := Call(ctx, s.LLMChain, llmInputs, opt...)
	if err != nil {
		return nil, err
	}
	sqlQuery, ok := out["text"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOutputValues, "text")
	}
	sqlQuery = strings.TrimSpace(sqlQuery)

	// Execute sql query
	queryResult, err := s.Database.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	// Generate answer
	llmInputs["input"] = query + queryPrefixWith + sqlQuery + stopWord + queryResult
	out, err = Call(ctx, s.LLMChain, llmInputs, options...)
	if err != nil {
		return nil, err
	}
	result, ok := out["text"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOutputValues, "text")
	}
	// Hack answer string
	strs := strings.Split(strings.Split(result, "\n\n")[0], "Answer:")
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
