package bedrockclient

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected string
	}{
		{
			name:     "ai21 provider",
			modelID:  "ai21.j2-ultra",
			expected: "ai21",
		},
		{
			name:     "amazon provider",
			modelID:  "amazon.titan-text-express-v1",
			expected: "amazon",
		},
		{
			name:     "anthropic provider",
			modelID:  "anthropic.claude-v2",
			expected: "anthropic",
		},
		{
			name:     "cohere provider",
			modelID:  "cohere.command-text-v14",
			expected: "cohere",
		},
		{
			name:     "meta provider",
			modelID:  "meta.llama2-13b-chat-v1",
			expected: "meta",
		},
		{
			name:     "unknown provider",
			modelID:  "unknown.model",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProvider(tt.modelID)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessInputMessagesGeneric(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		expected string
	}{
		{
			name:     "empty messages",
			messages: []Message{},
			expected: "",
		},
		{
			name: "single text message without role",
			messages: []Message{
				{Type: "text", Content: "Hello world"},
			},
			expected: "Hello world",
		},
		{
			name: "multiple messages with roles",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hi there"},
				{Role: llms.ChatMessageTypeAI, Type: "text", Content: "Hello! How can I help?"},
			},
			expected: "\nhuman: Hi there\nai: Hello! How can I help?\nAI: ",
		},
		{
			name: "mixed messages with and without roles",
			messages: []Message{
				{Type: "text", Content: "Start"},
				{Role: llms.ChatMessageTypeSystem, Type: "text", Content: "Be helpful"},
			},
			expected: "Start\nsystem: Be helpful\nAI: ",
		},
		{
			name: "non-text messages are ignored",
			messages: []Message{
				{Type: "text", Content: "Text message"},
				{Type: "image", Content: "image data"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Another text"},
			},
			expected: "Text message\nhuman: Another text\nAI: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processInputMessagesGeneric(tt.messages)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMaxTokens(t *testing.T) {
	tests := []struct {
		name         string
		maxTokens    int
		defaultValue int
		expected     int
	}{
		{
			name:         "positive max tokens",
			maxTokens:    100,
			defaultValue: 50,
			expected:     100,
		},
		{
			name:         "zero max tokens uses default",
			maxTokens:    0,
			defaultValue: 50,
			expected:     50,
		},
		{
			name:         "negative max tokens uses default",
			maxTokens:    -10,
			defaultValue: 50,
			expected:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxTokens(tt.maxTokens, tt.defaultValue)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateCompletion_UnsupportedProvider(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	require.NoError(t, err)

	bedrockClient := bedrockruntime.NewFromConfig(cfg)
	client := NewClient(bedrockClient)

	messages := []Message{
		{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hello"},
	}
	options := llms.CallOptions{}

	_, err = client.CreateCompletion(context.Background(), "unsupported.model", messages, options)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported provider")
}

// AI21 provider tests
func TestCreateAi21Completion_RequestStructure(t *testing.T) {
	messages := []Message{
		{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "What is the capital of France?"},
	}
	options := llms.CallOptions{
		Temperature:       0.7,
		TopP:              0.9,
		MaxTokens:         100,
		StopWords:         []string{"END"},
		RepetitionPenalty: 1.2,
		CandidateCount:    2,
	}

	// Create the input that would be sent to AI21
	txt := processInputMessagesGeneric(messages)
	input := ai21TextGenerationInput{
		Prompt:        txt,
		Temperature:   options.Temperature,
		TopP:          options.TopP,
		MaxTokens:     getMaxTokens(options.MaxTokens, 2048),
		StopSequences: options.StopWords,
		CountPenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: options.RepetitionPenalty},
		PresencePenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: 0},
		FrequencyPenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: 0},
		NumResults: options.CandidateCount,
	}

	// Verify the request structure
	body, err := json.Marshal(input)
	require.NoError(t, err)

	var unmarshaled ai21TextGenerationInput
	err = json.Unmarshal(body, &unmarshaled)
	require.NoError(t, err)

	require.Equal(t, "\nhuman: What is the capital of France?\nAI: ", unmarshaled.Prompt)
	require.Equal(t, 0.7, unmarshaled.Temperature)
	require.Equal(t, 0.9, unmarshaled.TopP)
	require.Equal(t, 100, unmarshaled.MaxTokens)
	require.Equal(t, []string{"END"}, unmarshaled.StopSequences)
	require.Equal(t, 1.2, unmarshaled.CountPenalty.Scale)
	require.Equal(t, 2, unmarshaled.NumResults)
}

// Amazon provider tests
func TestCreateAmazonCompletion_RequestStructure(t *testing.T) {
	messages := []Message{
		{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Tell me about AI"},
	}
	options := llms.CallOptions{
		Temperature: 0.5,
		TopP:        0.8,
		MaxTokens:   150,
		StopWords:   []string{"|", "User:"},
	}

	txt := processInputMessagesGeneric(messages)
	input := amazonTextGenerationInput{
		InputText: txt,
		TextGenerationConfig: amazonTextGenerationConfigInput{
			MaxTokens:     getMaxTokens(options.MaxTokens, 512),
			TopP:          options.TopP,
			Temperature:   options.Temperature,
			StopSequences: options.StopWords,
		},
	}

	body, err := json.Marshal(input)
	require.NoError(t, err)

	var unmarshaled amazonTextGenerationInput
	err = json.Unmarshal(body, &unmarshaled)
	require.NoError(t, err)

	require.Equal(t, "\nhuman: Tell me about AI\nAI: ", unmarshaled.InputText)
	require.Equal(t, 150, unmarshaled.TextGenerationConfig.MaxTokens)
	require.Equal(t, 0.8, unmarshaled.TextGenerationConfig.TopP)
	require.Equal(t, 0.5, unmarshaled.TextGenerationConfig.Temperature)
	require.Equal(t, []string{"|", "User:"}, unmarshaled.TextGenerationConfig.StopSequences)
}

// Anthropic provider tests
func TestProcessInputMessagesAnthropic(t *testing.T) {
	tests := []struct {
		name           string
		messages       []Message
		expectedMsgs   int
		expectedSystem string
		expectError    bool
		errorContains  string
	}{
		{
			name: "simple user message",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hello"},
			},
			expectedMsgs:   1,
			expectedSystem: "",
		},
		{
			name: "system message extracted",
			messages: []Message{
				{Role: llms.ChatMessageTypeSystem, Type: "text", Content: "You are helpful"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hi"},
			},
			expectedMsgs:   1,
			expectedSystem: "You are helpful",
		},
		{
			name: "multiple system messages concatenated",
			messages: []Message{
				{Role: llms.ChatMessageTypeSystem, Type: "text", Content: "First system"},
				{Role: llms.ChatMessageTypeSystem, Type: "text", Content: " Second system"},
			},
			expectedMsgs:   0,
			expectedSystem: "First system Second system",
		},
		{
			name: "system message with non-text content",
			messages: []Message{
				{Role: llms.ChatMessageTypeSystem, Type: "image", Content: "image data"},
			},
			expectError:   true,
			errorContains: "system prompt must be text",
		},
		{
			name: "conversation with alternating roles",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hello"},
				{Role: llms.ChatMessageTypeAI, Type: "text", Content: "Hi there"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "How are you?"},
			},
			expectedMsgs:   3,
			expectedSystem: "",
		},
		{
			name: "multiple messages same role chunked together",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "First"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Second"},
				{Role: llms.ChatMessageTypeAI, Type: "text", Content: "Response"},
			},
			expectedMsgs:   2,
			expectedSystem: "",
		},
		{
			name: "function role converted to user",
			messages: []Message{
				{Role: llms.ChatMessageTypeFunction, Type: "text", Content: "Function call"},
			},
			expectedMsgs:   1,
			expectedSystem: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs, system, err := processInputMessagesAnthropic(tt.messages)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				require.Len(t, msgs, tt.expectedMsgs)
				require.Equal(t, tt.expectedSystem, system)
			}
		})
	}
}

func TestGetAnthropicRole(t *testing.T) {
	tests := []struct {
		name        string
		role        llms.ChatMessageType
		expected    string
		expectError bool
	}{
		{
			name:     "system role",
			role:     llms.ChatMessageTypeSystem,
			expected: AnthropicSystem,
		},
		{
			name:     "AI role",
			role:     llms.ChatMessageTypeAI,
			expected: AnthropicRoleAssistant,
		},
		{
			name:     "human role",
			role:     llms.ChatMessageTypeHuman,
			expected: AnthropicRoleUser,
		},
		{
			name:     "generic role",
			role:     llms.ChatMessageTypeGeneric,
			expected: AnthropicRoleUser,
		},
		{
			name:     "function role treated as user",
			role:     llms.ChatMessageTypeFunction,
			expected: AnthropicRoleUser,
		},
		{
			name:     "tool role treated as user",
			role:     llms.ChatMessageTypeTool,
			expected: AnthropicRoleUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getAnthropicRole(tt.role)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetAnthropicInputContent(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		validate func(t *testing.T, content anthropicTextGenerationInputContent)
	}{
		{
			name: "text message",
			message: Message{
				Type:    AnthropicMessageTypeText,
				Content: "Hello world",
			},
			validate: func(t *testing.T, content anthropicTextGenerationInputContent) {
				require.Equal(t, AnthropicMessageTypeText, content.Type)
				require.Equal(t, "Hello world", content.Text)
				require.Nil(t, content.Source)
			},
		},
		{
			name: "image message",
			message: Message{
				Type:     AnthropicMessageTypeImage,
				Content:  "image data",
				MimeType: "image/png",
			},
			validate: func(t *testing.T, content anthropicTextGenerationInputContent) {
				require.Equal(t, AnthropicMessageTypeImage, content.Type)
				require.Empty(t, content.Text)
				require.NotNil(t, content.Source)
				require.Equal(t, "base64", content.Source.Type)
				require.Equal(t, "image/png", content.Source.MediaType)
				require.NotEmpty(t, content.Source.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAnthropicInputContent(tt.message)
			tt.validate(t, result)
		})
	}
}

// Cohere provider tests
func TestCreateCohereCompletion_RequestStructure(t *testing.T) {
	messages := []Message{
		{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Explain quantum computing"},
	}
	options := llms.CallOptions{
		Temperature:    0.8,
		TopP:           0.95,
		TopK:           50,
		MaxTokens:      200,
		StopWords:      []string{"END", "STOP"},
		CandidateCount: 3,
	}

	txt := processInputMessagesGeneric(messages)
	input := &cohereTextGenerationInput{
		Prompt:         txt,
		Temperature:    options.Temperature,
		P:              options.TopP,
		K:              options.TopK,
		MaxTokens:      getMaxTokens(options.MaxTokens, 20),
		StopSequences:  options.StopWords,
		NumGenerations: options.CandidateCount,
	}

	body, err := json.Marshal(input)
	require.NoError(t, err)

	var unmarshaled cohereTextGenerationInput
	err = json.Unmarshal(body, &unmarshaled)
	require.NoError(t, err)

	require.Equal(t, "\nhuman: Explain quantum computing\nAI: ", unmarshaled.Prompt)
	require.Equal(t, 0.8, unmarshaled.Temperature)
	require.Equal(t, 0.95, unmarshaled.P)
	require.Equal(t, 50, unmarshaled.K)
	require.Equal(t, 200, unmarshaled.MaxTokens)
	require.Equal(t, []string{"END", "STOP"}, unmarshaled.StopSequences)
	require.Equal(t, 3, unmarshaled.NumGenerations)
}

// Meta provider tests
func TestCreateMetaCompletion_RequestStructure(t *testing.T) {
	messages := []Message{
		{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Write a poem about technology"},
	}
	options := llms.CallOptions{
		Temperature: 0.6,
		TopP:        0.85,
		MaxTokens:   256,
	}

	txt := processInputMessagesGeneric(messages)
	input := &metaTextGenerationInput{
		Prompt:      txt,
		Temperature: options.Temperature,
		TopP:        options.TopP,
		MaxGenLen:   getMaxTokens(options.MaxTokens, 512),
	}

	body, err := json.Marshal(input)
	require.NoError(t, err)

	var unmarshaled metaTextGenerationInput
	err = json.Unmarshal(body, &unmarshaled)
	require.NoError(t, err)

	require.Equal(t, "\nhuman: Write a poem about technology\nAI: ", unmarshaled.Prompt)
	require.Equal(t, 0.6, unmarshaled.Temperature)
	require.Equal(t, 0.85, unmarshaled.TopP)
	require.Equal(t, 256, unmarshaled.MaxGenLen)
}

// Response parsing tests
func TestAi21ResponseParsing(t *testing.T) {
	output := ai21TextGenerationOutput{
		ID: "12345",
		Prompt: struct {
			Tokens []struct{} `json:"tokens"`
		}{
			Tokens: make([]struct{}, 10), // 10 input tokens
		},
		Completions: []struct {
			Data struct {
				Text   string     `json:"text"`
				Tokens []struct{} `json:"tokens"`
			} `json:"data"`
			FinishReason struct {
				Reason string `json:"reason"`
			} `json:"finishReason"`
		}{
			{
				Data: struct {
					Text   string     `json:"text"`
					Tokens []struct{} `json:"tokens"`
				}{
					Text:   "Paris is the capital of France.",
					Tokens: make([]struct{}, 8), // 8 output tokens
				},
				FinishReason: struct {
					Reason string `json:"reason"`
				}{
					Reason: Ai21CompletionReasonStop,
				},
			},
			{
				Data: struct {
					Text   string     `json:"text"`
					Tokens []struct{} `json:"tokens"`
				}{
					Text:   "The capital of France is Paris.",
					Tokens: make([]struct{}, 8), // 8 output tokens
				},
				FinishReason: struct {
					Reason string `json:"reason"`
				}{
					Reason: Ai21CompletionReasonEndOfText,
				},
			},
		},
	}

	// Test that we can marshal and unmarshal properly
	data, err := json.Marshal(output)
	require.NoError(t, err)

	var parsed ai21TextGenerationOutput
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	require.Equal(t, output.ID, parsed.ID)
	require.Len(t, parsed.Completions, 2)
	require.Equal(t, "Paris is the capital of France.", parsed.Completions[0].Data.Text)
	require.Equal(t, Ai21CompletionReasonStop, parsed.Completions[0].FinishReason.Reason)
}

func TestAmazonResponseParsing(t *testing.T) {
	output := amazonTextGenerationOutput{
		InputTextTokenCount: 5,
		Results: []struct {
			TokenCount       int    `json:"tokenCount"`
			OutputText       string `json:"outputText"`
			CompletionReason string `json:"completionReason"`
		}{
			{
				TokenCount:       15,
				OutputText:       "AI is transforming the world.",
				CompletionReason: AmazonCompletionReasonFinish,
			},
		},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var parsed amazonTextGenerationOutput
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	require.Equal(t, 5, parsed.InputTextTokenCount)
	require.Len(t, parsed.Results, 1)
	require.Equal(t, 15, parsed.Results[0].TokenCount)
	require.Equal(t, "AI is transforming the world.", parsed.Results[0].OutputText)
	require.Equal(t, AmazonCompletionReasonFinish, parsed.Results[0].CompletionReason)
}

func TestCohereResponseParsing(t *testing.T) {
	output := cohereTextGenerationOutput{
		ID: "cohere-12345",
		Generations: []*cohereTextGenerationOutputGeneration{
			{
				ID:           "gen-1",
				Index:        0,
				FinishReason: CohereCompletionReasonComplete,
				Text:         "Quantum computing uses quantum mechanics principles.",
			},
			{
				ID:           "gen-2",
				Index:        1,
				FinishReason: CohereCompletionReasonMaxTokens,
				Text:         "Quantum computing is a revolutionary technology that...",
			},
		},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var parsed cohereTextGenerationOutput
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	require.Equal(t, "cohere-12345", parsed.ID)
	require.Len(t, parsed.Generations, 2)
	require.Equal(t, "gen-1", parsed.Generations[0].ID)
	require.Equal(t, CohereCompletionReasonComplete, parsed.Generations[0].FinishReason)
}

func TestMetaResponseParsing(t *testing.T) {
	output := metaTextGenerationOutput{
		Generation:           "Technology advances at rapid pace,\nInnovation fills every space.",
		PromptTokenCount:     7,
		GenerationTokenCount: 12,
		StopReason:           MetaCompletionReasonStop,
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var parsed metaTextGenerationOutput
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	require.Equal(t, output.Generation, parsed.Generation)
	require.Equal(t, 7, parsed.PromptTokenCount)
	require.Equal(t, 12, parsed.GenerationTokenCount)
	require.Equal(t, MetaCompletionReasonStop, parsed.StopReason)
}

func TestAnthropicResponseParsing(t *testing.T) {
	output := anthropicTextGenerationOutput{
		Type: "message",
		Role: "assistant",
		Content: []anthropicContentBlock{
			{
				Type: "text",
				Text: "Hello! I'm Claude, an AI assistant.",
			},
		},
		StopReason:   AnthropicCompletionReasonEndTurn,
		StopSequence: "",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var parsed anthropicTextGenerationOutput
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	require.Equal(t, "message", parsed.Type)
	require.Equal(t, "assistant", parsed.Role)
	require.Len(t, parsed.Content, 1)
	require.Equal(t, "text", parsed.Content[0].Type)
	require.Equal(t, "Hello! I'm Claude, an AI assistant.", parsed.Content[0].Text)
	require.Equal(t, AnthropicCompletionReasonEndTurn, parsed.StopReason)
	require.Equal(t, 10, parsed.Usage.InputTokens)
	require.Equal(t, 15, parsed.Usage.OutputTokens)
}

// Edge case tests
func TestEmptyResponses(t *testing.T) {
	t.Run("Amazon empty results", func(t *testing.T) {
		output := amazonTextGenerationOutput{
			InputTextTokenCount: 5,
			Results: []struct {
				TokenCount       int    `json:"tokenCount"`
				OutputText       string `json:"outputText"`
				CompletionReason string `json:"completionReason"`
			}{},
		}

		// This should be handled as an error in the actual implementation
		require.Empty(t, output.Results)
	})

	t.Run("Anthropic empty content", func(t *testing.T) {
		output := anthropicTextGenerationOutput{
			Type:       "message",
			Role:       "assistant",
			Content:    []anthropicContentBlock{},
			StopReason: AnthropicCompletionReasonEndTurn,
		}

		// This should be handled as an error in the actual implementation
		require.Empty(t, output.Content)
	})
}

// Test streaming response parsing for Anthropic
func TestAnthropicStreamingResponseChunk(t *testing.T) {
	tests := []struct {
		name  string
		chunk streamingCompletionResponseChunk
	}{
		{
			name: "message_start chunk",
			chunk: streamingCompletionResponseChunk{
				Type: "message_start",
				Message: struct {
					ID           string `json:"id"`
					Type         string `json:"type"`
					Role         string `json:"role"`
					Content      []any  `json:"content"`
					Model        string `json:"model"`
					StopReason   any    `json:"stop_reason"`
					StopSequence any    `json:"stop_sequence"`
					Usage        struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					} `json:"usage"`
				}{
					ID:    "msg-123",
					Type:  "message",
					Role:  "assistant",
					Model: "claude-3",
					Usage: struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					}{
						InputTokens: 25,
					},
				},
			},
		},
		{
			name: "content_block_delta chunk",
			chunk: streamingCompletionResponseChunk{
				Type:  "content_block_delta",
				Index: 0,
				Delta: struct {
					Type         string `json:"type"`
					Text         string `json:"text"`
					StopReason   string `json:"stop_reason"`
					StopSequence any    `json:"stop_sequence"`
				}{
					Type: "text_delta",
					Text: "Hello, how can I help you today?",
				},
			},
		},
		{
			name: "message_delta chunk",
			chunk: streamingCompletionResponseChunk{
				Type: "message_delta",
				Delta: struct {
					Type         string `json:"type"`
					Text         string `json:"text"`
					StopReason   string `json:"stop_reason"`
					StopSequence any    `json:"stop_sequence"`
				}{
					StopReason: AnthropicCompletionReasonEndTurn,
				},
				Usage: struct {
					OutputTokens int `json:"output_tokens"`
				}{
					OutputTokens: 12,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.chunk)
			require.NoError(t, err)

			var parsed streamingCompletionResponseChunk
			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)

			require.Equal(t, tt.chunk.Type, parsed.Type)

			switch tt.chunk.Type {
			case "message_start":
				require.Equal(t, tt.chunk.Message.ID, parsed.Message.ID)
				require.Equal(t, tt.chunk.Message.Usage.InputTokens, parsed.Message.Usage.InputTokens)
			case "content_block_delta":
				require.Equal(t, tt.chunk.Delta.Text, parsed.Delta.Text)
			case "message_delta":
				require.Equal(t, tt.chunk.Delta.StopReason, parsed.Delta.StopReason)
				require.Equal(t, tt.chunk.Usage.OutputTokens, parsed.Usage.OutputTokens)
			}
		})
	}
}

// Test AWS SDK request structures
func TestBedrockRequestStructures(t *testing.T) {
	t.Run("InvokeModelInput", func(t *testing.T) {
		input := &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String("anthropic.claude-v2"),
			Accept:      aws.String("*/*"),
			ContentType: aws.String("application/json"),
			Body:        []byte(`{"prompt": "Hello"}`),
		}

		require.Equal(t, "anthropic.claude-v2", *input.ModelId)
		require.Equal(t, "*/*", *input.Accept)
		require.Equal(t, "application/json", *input.ContentType)
		require.Equal(t, []byte(`{"prompt": "Hello"}`), input.Body)
	})

	t.Run("InvokeModelWithResponseStreamInput", func(t *testing.T) {
		input := &bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String("anthropic.claude-3-sonnet"),
			Accept:      aws.String("*/*"),
			ContentType: aws.String("application/json"),
			Body:        []byte(`{"prompt": "Stream this"}`),
		}

		require.Equal(t, "anthropic.claude-3-sonnet", *input.ModelId)
		require.Equal(t, "*/*", *input.Accept)
		require.Equal(t, "application/json", *input.ContentType)
		require.Equal(t, []byte(`{"prompt": "Stream this"}`), input.Body)
	})
}
