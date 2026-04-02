package bedrockclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/tmc/langchaingo/llms"
)

const bedrockGuardrailIntervened = "INTERVENED"

type bedrockGuardrailParams struct {
	identifier           *string
	version              *string
	tagSuffix            *string
	streamProcessingMode *string
}

func getRequiredGuardrailParam(safety map[string]any, key string) (*string, error) {
	value, ok := safety[key]
	if !ok {
		return nil, fmt.Errorf("bedrock guardrail %s is required when safety config is provided", key)
	}

	str, ok := value.(string)
	if !ok || strings.TrimSpace(str) == "" {
		return nil, fmt.Errorf("bedrock guardrail %s must be a non-empty string", key)
	}

	return aws.String(str), nil
}

func getOptionalGuardrailParam(safety map[string]any, key string) (*string, error) {
	value, ok := safety[key]
	if !ok {
		return nil, nil
	}

	str, ok := value.(string)
	if !ok || strings.TrimSpace(str) == "" {
		return nil, fmt.Errorf("bedrock guardrail %s must be a non-empty string if provided", key)
	}

	return aws.String(str), nil
}

func getGuardrailParams(options llms.CallOptions) (bedrockGuardrailParams, error) {
	if options.SafetyConfig == nil {
		return bedrockGuardrailParams{}, nil
	}

	identifier, err := getRequiredGuardrailParam(options.SafetyConfig, "identifier")
	if err != nil {
		return bedrockGuardrailParams{}, err
	}
	version, err := getRequiredGuardrailParam(options.SafetyConfig, "version")
	if err != nil {
		return bedrockGuardrailParams{}, err
	}

	tagSuffix, err := getOptionalGuardrailParam(options.SafetyConfig, "tagSuffix")
	if err != nil {
		return bedrockGuardrailParams{}, err
	}
	streamProcessingMode, err := getOptionalGuardrailParam(options.SafetyConfig, "streamProcessingMode")
	if err != nil {
		return bedrockGuardrailParams{}, err
	}

	return bedrockGuardrailParams{
		identifier:           identifier,
		version:              version,
		tagSuffix:            tagSuffix,
		streamProcessingMode: streamProcessingMode,
	}, nil
}

func handleGuardrailParams(input guardrailConfigurableAdapter, options llms.CallOptions) error {
	guardrailParams, err := getGuardrailParams(options)
	if err != nil {
		return fmt.Errorf("failed to get guardrail parameters: %w", err)
	}

	if guardrailParams.identifier != nil && guardrailParams.version != nil {
		input.setGuardrailIdentifier(guardrailParams.identifier)
		input.setGuardrailVersion(guardrailParams.version)
	}

	if guardrailParams.tagSuffix != nil || guardrailParams.streamProcessingMode != nil {
		var m map[string]any
		if err := json.Unmarshal(input.body(), &m); err != nil {
			return fmt.Errorf("failed to unmarshal input body when adding guardrail parameters: %w", err)
		}

		var config = make(map[string]string)
		if guardrailParams.tagSuffix != nil {
			config["tagSuffix"] = *guardrailParams.tagSuffix
		}
		if guardrailParams.streamProcessingMode != nil {
			config["streamProcessingMode"] = *guardrailParams.streamProcessingMode
		}
		m["amazon-bedrock-guardrailConfig"] = config

		newBody, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("failed to marshal input body when adding guardrail parameters: %w", err)
		}
		input.setBody(newBody)
	}
	return nil
}

func checkGuardrailError(response any) error {
	var action string
	var trace json.RawMessage

	switch r := response.(type) {
	case []byte:
		var resp guardrailResponse
		if err := json.Unmarshal(r, &resp); err != nil {
			// If we can't parse the response, we assume it's not a guardrail error and return nil
			return nil
		}
		action = resp.AmazonBedrockGuardrailAction
		trace = resp.AmazonBedrockTrace
	case streamingCompletionResponseChunk:
		action = r.AmazonBedrockGuardrailAction
		trace = r.AmazonBedrockTrace
	case *streamingCompletionResponseChunk:
		action = r.AmazonBedrockGuardrailAction
		trace = r.AmazonBedrockTrace
	default:
		return nil
	}

	if action == bedrockGuardrailIntervened {
		err := llms.NewError(llms.ErrCodeContentFilter, "bedrock", "content blocked by Bedrock guardrail")
		err.WithDetail("guardrail_action", action)
		if len(trace) > 0 {
			err.WithDetail("trace", string(trace))
		}
		return err
	}
	return nil
}
