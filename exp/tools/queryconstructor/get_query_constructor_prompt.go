package queryconstructor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	queryconstructor_prompts "github.com/tmc/langchaingo/exp/tools/queryconstructor/prompts"
	"github.com/tmc/langchaingo/prompts"
)

//go:embed prompts/query.txt
var _query string //nolint:gochecknoglobals
//go:embed prompts/default_prefix.txt
var _defaultPrefixPrompt string //nolint:gochecknoglobals
//go:embed prompts/default_suffix.txt
var _defaultSuffixPrompt string //nolint:gochecknoglobals
//go:embed prompts/default_schema.txt
var _defaultSchemaPrompt string //nolint:gochecknoglobals
//go:embed prompts/suffix_without_datasource.txt
var _suffixWithoutDatasourcePrompt string //nolint:gochecknoglobals
//go:embed prompts/schema_with_limit.txt
var _schemaWithLimitPrompt string //nolint:gochecknoglobals
//go:embed prompts/user_specified_example.txt
var _userSpecifiedExample string //nolint:gochecknoglobals
//go:embed prompts/with_data_source.txt
var _withDataSource string //nolint:gochecknoglobals
//go:embed prompts/example_prompt.txt
var _examplePrompt string //nolint:gochecknoglobals

// documentContents: The contents of the document to be queried.
// attributeInfo: A list of AttributeInfo objects describing the attributes of the document.
// examples: Optional list of examples to use for the chain.
// allowedComparators: Sequence of allowed comparators.
// allowedOperators: Sequence of allowed operators.
// schemaPrompt: Prompt for describing query schema. Should have string input
//             variables allowed_comparators and allowed_operators.
// enableLimit: Whether to enable the limit operator. Defaults to False.

type GetQueryConstructorPromptArgs struct {
	DocumentContents   string
	AttributeInfo      []AttributeInfo
	AllowedComparators []Comparator
	AllowedOperators   []Operator
	InputOuputExamples []InputOuputExample
	CustomExamples     []map[string]string
	SchemaPrompt       *string
	EnableLimit        *bool
}

// Create query construction prompt.
func GetQueryConstructorPrompt(args GetQueryConstructorPromptArgs) (*prompts.FewShotPrompt, error) {

	defaultSchema := getDefaultSchema(args.SchemaPrompt, args.EnableLimit)

	schema, err := prompts.NewPromptTemplate(defaultSchema, []string{
		"allowed_comparators",
		"allowed_operators",
	}).Format(map[string]any{
		"allowed_comparators": strings.Join(args.AllowedComparators, " | "),
		"allowed_operators":   strings.Join(args.AllowedOperators, " | "),
	})
	if err != nil {
		return nil, fmt.Errorf("error formating 'default schema' prompt %w", err)
	}

	jsonAttributes, err := formatAttribute(args.AttributeInfo)
	if err != nil {
		return nil, fmt.Errorf("error formating attribute while loading constructor %w", err)
	}

	outputExample := setExampleOutput{}

	if args.InputOuputExamples != nil && len(args.InputOuputExamples) > 0 {
		if err := setInputOutputExamples(setInputOutputExamplesInput{
			examples:         args.InputOuputExamples,
			schema:           schema,
			jsonAttributes:   string(jsonAttributes),
			documentContents: args.DocumentContents,
			enableLimit:      args.EnableLimit,
		}, &outputExample); err != nil {
			return nil, fmt.Errorf("error setting input output example %w", err)
		}
	} else {
		examples := getDefaultExamples(args.CustomExamples, args.EnableLimit)
		outputExample.examples = examples
		if err := setExamples(setExamplesInput{
			examples:         examples,
			schema:           schema,
			jsonAttributes:   string(jsonAttributes),
			documentContents: args.DocumentContents,
			enableLimit:      args.EnableLimit,
		}, &outputExample); err != nil {
			return nil, fmt.Errorf("error setting custom example %w", err)
		}
	}

	outputExample.suffix = outputExample.suffix + _query

	return prompts.NewFewShotPrompt(outputExample.examplePrompt, outputExample.examples, nil, outputExample.prefix, outputExample.suffix, []string{"query"}, nil, "", prompts.TemplateFormatGoTemplate, true)
}

func getDefaultSchema(schemaPrompt *string, enableLimit *bool) string {
	if schemaPrompt != nil {
		return *schemaPrompt
	}

	if enableLimit != nil && *enableLimit {
		return _schemaWithLimitPrompt
	}

	return _defaultSchemaPrompt
}

func getDefaultExamples(customExamples []map[string]string, enableLimit *bool) []map[string]string {
	if customExamples != nil {
		return customExamples
	}

	if enableLimit != nil && *enableLimit {
		return queryconstructor_prompts.ExamplesWithLimit
	}

	return queryconstructor_prompts.DefaultExamples
}

func formatAttribute(attributeInfo []AttributeInfo) ([]byte, error) {
	output := map[string]map[string]interface{}{}
	for _, ai := range attributeInfo {
		output[ai.Name] = map[string]interface{}{
			"description": ai.Description,
			"type":        ai.Type,
		}
	}

	return json.Marshal(output)
}

type setExampleOutput struct {
	examplePrompt prompts.PromptTemplate
	examples      []map[string]string
	prefix        string
	suffix        string
}

type setInputOutputExamplesInput struct {
	examples         []InputOuputExample
	schema           string
	jsonAttributes   string
	documentContents string
	enableLimit      *bool
}

func setInputOutputExamples(input setInputOutputExamplesInput, output *setExampleOutput) error {
	formattedExamples := []map[string]string{}
	var err error
	for i, e := range input.examples {
		structuredQuery, err := json.Marshal(e.Ouput)
		if err != nil {
			return fmt.Errorf("error marshalling output of example %w", err)
		}

		formattedExamples = append(formattedExamples, map[string]string{
			"i":                strconv.Itoa(i),
			"user_query":       e.Input,
			"structured_query": string(structuredQuery),
		})

	}
	output.examples = formattedExamples

	output.examplePrompt = prompts.NewPromptTemplate(_userSpecifiedExample, []string{"i", "user_query", "structured_request"})

	if output.prefix, err = prompts.NewPromptTemplate(_defaultPrefixPrompt+_withDataSource, []string{
		"schema",
		"content",
		"attributes",
	}).Format(map[string]interface{}{
		"schema":     input.schema,
		"content":    input.documentContents,
		"attributes": input.jsonAttributes,
	}); err != nil {
		return fmt.Errorf("error formating 'default prefix' and 'with data source' prompt %w", err)
	}

	if output.suffix, err = prompts.NewPromptTemplate(_suffixWithoutDatasourcePrompt, []string{"i"}).Format(map[string]interface{}{
		"i": strconv.Itoa(len(input.examples) + 1),
	}); err != nil {
		return fmt.Errorf("error formating 'suffix without data source' prompt %w", err)
	}

	return nil
}

type setExamplesInput struct {
	examples         []map[string]string
	schema           string
	jsonAttributes   string
	documentContents string
	enableLimit      *bool
}

func setExamples(input setExamplesInput, output *setExampleOutput) error {
	var err error
	output.examplePrompt = prompts.NewPromptTemplate(_examplePrompt, []string{
		"i",
		"data_source",
		"user_query",
		"structured_request",
	})

	if output.prefix, err = prompts.NewPromptTemplate(_defaultPrefixPrompt, []string{"schema"}).Format(map[string]any{
		"schema": input.schema,
	}); err != nil {
		return fmt.Errorf("error formating 'default prefix' prompt %w", err)
	}

	if output.suffix, err = prompts.NewPromptTemplate(_defaultSuffixPrompt, []string{"i", "content", "attributes"}).Format(map[string]any{
		"i":          strconv.Itoa(len(input.examples) + 1),
		"content":    input.documentContents,
		"attributes": input.jsonAttributes,
	}); err != nil {
		return fmt.Errorf("error formating 'default suffix' prompt %w", err)
	}

	return nil
}
