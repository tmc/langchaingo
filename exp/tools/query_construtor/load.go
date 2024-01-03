package queryconstrutor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	queryconstructor_prompts "github.com/tmc/langchaingo/exp/tools/query_construtor/prompts"
	"github.com/tmc/langchaingo/prompts"
)

//go:embed prompts/default_prefix.txt
var _defaultPrefixPrompt string //nolint:gochecknoglobals
//go:embed prompts/default_schema.txt
var _defaultSchemaPrompt string //nolint:gochecknoglobals
//go:embed prompts/suffix_without_datasource.txt
var _suffixWithoutDatasourcePrompt string //nolint:gochecknoglobals
//go:embed prompts/default_suffix.txt
var _defaultSuffixPrompt string //nolint:gochecknoglobals
//go:embed prompts/schema_with_limit.txt
var _schemaWithLimitPrompt string //nolint:gochecknoglobals
//go:embed prompts/user_specified_example.txt
var _userSpecifiedExample string //nolint:gochecknoglobals
//go:embed prompts/with_data_source.txt
var _withDataSource string //nolint:gochecknoglobals

// documentContents: The contents of the document to be queried.
// attributeInfo: A list of AttributeInfo objects describing the attributes of the document.
// examples: Optional list of examples to use for the chain.
// allowedComparators: Sequence of allowed comparators.
// allowedOperators: Sequence of allowed operators.
// schemaPrompt: Prompt for describing query schema. Should have string input
//             variables allowed_comparators and allowed_operators.
// enableLimit: Whether to enable the limit operator. Defaults to False.

type LoadArgs struct {
	documentContents   string
	attributeInfo      []AttributeInfo
	allowedComparators []Comparator
	allowedOperators   []Operator
	examples           []Example
	customExamples     []map[string]string
	schemaPrompt       *string
	enableLimit        *bool
}

// Create query construction prompt.
func Load(args LoadArgs) (*prompts.FewShotPrompt, error) {
	defaultSchema := getDefaultSchema(args.schemaPrompt, args.enableLimit)

	schema, err := prompts.NewPromptTemplate(defaultSchema, []string{
		"allowed_comparators",
		"allowed_operators",
	}).Format(map[string]any{
		"allowed_comparators": strings.Join(args.allowedComparators, " | "),
		"allowed_operators":   strings.Join(args.allowedOperators, " | "),
	})

	jsonAttributes, err := formatAttribute(args.attributeInfo)
	if err != nil {
		fmt.Errorf("error formating attribute while loading constructor %w", err)
	}

	outputExample := setExampleOutput{}

	if args.examples != nil && len(args.examples) > 0 {
		if err := setInputOutputExamples(setInputOutputExampleInput{
			examples:         args.examples,
			schema:           schema,
			jsonAttributes:   string(jsonAttributes),
			documentContents: args.documentContents,
			enableLimit:      args.enableLimit,
		}, &outputExample); err != nil {
			return nil, fmt.Errorf("error setting input output example %w", err)
		}
	}

	if args.customExamples != nil && len(args.customExamples) > 0 {

	}

	examplePrompt, examples, prefix, suffix, err := setExamples(args.examples, args.customExamples, schema, string(jsonAttributes), args.documentContents, args.enableLimit)
	if err != nil {
		return nil, fmt.Errorf("error setting examples for query constructor %w", err)
	}

	return prompts.NewFewShotPrompt(*examplePrompt, examples, nil, *prefix, *suffix, []string{""}, nil, "", prompts.TemplateFormatGoTemplate, true)
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
	var output map[string]map[string]interface{}
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
	examples      []Example
	prefix        string
	suffix        string
}

type setInputOutputExampleInput struct {
	examples         []Example
	schema           string
	jsonAttributes   string
	documentContents string
	enableLimit      *bool
}

func setInputOutputExamples(input setInputOutputExampleInput, output *setExampleOutput) error {
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
		return fmt.Errorf("error formating 'default prefix' prompt %w", err)
	}

	if output.suffix, err = prompts.NewPromptTemplate(_suffixWithoutDatasourcePrompt, []string{"i"}).Format(map[string]interface{}{
		"i": strconv.Itoa(len(input.examples) + 1),
	}); err != nil {
		return fmt.Errorf("error formating 'suffix without data source' prompt %w", err)
	}

	return nil
}

func setCustomExamples(examples *[]Example, customExamples []map[string]string, schema string, jsonAttributes string, documentContents string, enableLimit *bool) (*prompts.PromptTemplate, []map[string]string, *string, *string, error) {

}
