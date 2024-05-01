package detectschema

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/chains"
	detectschemaprompts "github.com/tmc/langchaingo/exp/detectschema/prompts"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

//go:embed prompts/prefix.txt
var _prefix string //nolint:gochecknoglobals
//go:embed prompts/example_template.txt
var _exampleTemplate string //nolint:gochecknoglobals
//go:embed prompts/suffix.txt
var _suffix string //nolint:gochecknoglobals

func (d *Detector) GetAttributeInfo(ctx context.Context, fileName string, fileType string, sampleData string) ([]schema.AttributeInfo, error) {
	exampleTemplatePrompt := prompts.NewPromptTemplate(_exampleTemplate, []string{
		"i",
		"file_type",
		"sample_data",
		"response",
	})

	prompt, err := prompts.NewFewShotPrompt(exampleTemplatePrompt, detectschemaprompts.GetExamples(), nil, _prefix, _suffix, []string{"file_name", "file_type", "sample_data", "types"}, nil, "", prompts.TemplateFormatGoTemplate, true)
	if err != nil {
		return nil, err
	}

	promptChain := *chains.NewLLMChain(
		d.llm,
		prompt,
		chains.WithTemperature(0),
	)

	promptChain.OutputParser = outputparser.NewJSONMarkdown()

	result, err := promptChain.Call(ctx, map[string]any{
		"file_name":   fileName,
		"file_type":   fileType,
		"sample_data": sampleData,
		"types":       strings.Join([]string{AllowedTypeBool, AllowedTypeFloat, AllowedTypeInt, AllowedTypeString}, ","),
	})
	if err != nil {
		return nil, err
	}

	output := []schema.AttributeInfo{}
	var resultBytes []byte
	var ok bool

	if resultBytes, ok = result["text"].([]byte); !ok {
		return nil, fmt.Errorf("wrong type retuned by json markdown parser")
	}

	if err = json.Unmarshal(resultBytes, &output); err != nil {
		return nil, fmt.Errorf("wrong json retuned by json markdown parser")
	}

	return output, nil
}
