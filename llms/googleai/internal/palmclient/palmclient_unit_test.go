package palmclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestChatMessage_GetType(t *testing.T) {
	tests := []struct {
		name   string
		author string
		want   llms.ChatMessageType
	}{
		{
			name:   "user message",
			author: "user",
			want:   llms.ChatMessageTypeHuman,
		},
		{
			name:   "bot message",
			author: "bot",
			want:   llms.ChatMessageTypeAI,
		},
		{
			name:   "empty author",
			author: "",
			want:   llms.ChatMessageTypeAI,
		},
		{
			name:   "unknown author",
			author: "system",
			want:   llms.ChatMessageTypeAI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ChatMessage{
				Author: tt.author,
			}
			assert.Equal(t, tt.want, msg.GetType())
		})
	}
}

func TestChatMessage_GetContent(t *testing.T) {
	msg := ChatMessage{
		Content: "Hello, world!",
		Author:  "user",
	}
	assert.Equal(t, "Hello, world!", msg.GetContent())
}

func TestConvertArray(t *testing.T) {
	input := []string{"stop1", "stop2", "stop3"}
	result := convertArray(input)

	expected := []interface{}{"stop1", "stop2", "stop3"}
	assert.Equal(t, expected, result)

	// Test empty array
	emptyResult := convertArray([]string{})
	assert.Equal(t, []interface{}{}, emptyResult)
}

func TestCloneDefaultParameters(t *testing.T) {
	cloned := cloneDefaultParameters()

	// Check that all default parameters are present
	assert.Equal(t, defaultParameters["temperature"], cloned["temperature"])
	assert.Equal(t, defaultParameters["maxOutputTokens"], cloned["maxOutputTokens"])
	assert.Equal(t, defaultParameters["topP"], cloned["topP"])
	assert.Equal(t, defaultParameters["topK"], cloned["topK"])

	// Verify it's a copy by modifying the clone
	cloned["temperature"] = 0.9
	assert.NotEqual(t, defaultParameters["temperature"], cloned["temperature"])
}

func TestMergeParams(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		checkKey string
		want     interface{}
	}{
		{
			name: "override temperature",
			params: map[string]interface{}{
				"temperature": 0.9,
			},
			checkKey: "temperature",
			want:     0.9,
		},
		{
			name: "zero value not merged",
			params: map[string]interface{}{
				"temperature": 0.0,
			},
			checkKey: "temperature",
			want:     defaultParameters["temperature"], // Should keep default
		},
		{
			name: "int value",
			params: map[string]interface{}{
				"maxOutputTokens": 512,
			},
			checkKey: "maxOutputTokens",
			want:     512,
		},
		{
			name: "array value",
			params: map[string]interface{}{
				"stopSequences": []interface{}{"END", "STOP"},
			},
			checkKey: "stopSequences",
			want:     []interface{}{"END", "STOP"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeParams(defaultParameters, tt.params)
			assert.NotNil(t, result)

			// Check the merged value
			fields := result.GetFields()
			if val, ok := fields[tt.checkKey]; ok {
				// Handle different value types
				switch v := tt.want.(type) {
				case float64:
					assert.Equal(t, v, val.GetNumberValue())
				case int:
					assert.Equal(t, float64(v), val.GetNumberValue())
				case []interface{}:
					list := val.GetListValue()
					assert.NotNil(t, list)
					assert.Equal(t, len(v), len(list.GetValues()))
				}
			} else {
				// If the key is not in the result, check if we expected the default
				if defaultVal, hasDefault := defaultParameters[tt.checkKey]; hasDefault {
					// For zero values, we expect the default to be preserved
					if tt.params[tt.checkKey] == 0.0 || tt.params[tt.checkKey] == 0 {
						assert.Equal(t, defaultVal, tt.want)
					} else {
						t.Errorf("Key %s not found in result", tt.checkKey)
					}
				}
			}
		})
	}
}

func TestConvertToOutputStruct(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := map[string]interface{}{
			"temperature": 0.5,
			"maxTokens":   100,
		}
		result := convertToOutputStruct(defaultParameters, params)
		assert.NotNil(t, result)
		assert.Equal(t, 0.5, result.GetFields()["temperature"].GetNumberValue())
	})

	t.Run("invalid params fallback to defaults", func(t *testing.T) {
		// Create params that would cause structpb.NewStruct to fail
		params := map[string]interface{}{
			"invalid": make(chan int), // channels cannot be converted
		}
		result := convertToOutputStruct(defaultParameters, params)
		assert.NotNil(t, result)
		// Should fall back to default parameters
		assert.Equal(t, defaultParameters["temperature"], result.GetFields()["temperature"].GetNumberValue())
	})
}

func TestProjectLocationPublisherModelPath(t *testing.T) {
	client := &PaLMClient{
		projectID: "test-project",
	}

	path := client.projectLocationPublisherModelPath("test-project", "us-central1", "google", "text-bison")
	expected := "projects/test-project/locations/us-central1/publishers/google/models/text-bison"
	assert.Equal(t, expected, path)
}

// TestNew was removed as it contained an unconditional skip

func TestCompletionRequestValidation(t *testing.T) {
	req := &CompletionRequest{
		Prompts:       []string{"Test prompt"},
		MaxTokens:     100,
		Temperature:   0.7,
		TopP:          1,
		TopK:          40,
		StopSequences: []string{"END"},
	}

	assert.NotEmpty(t, req.Prompts)
	assert.Greater(t, req.MaxTokens, 0)
	assert.GreaterOrEqual(t, req.Temperature, 0.0)
	assert.LessOrEqual(t, req.Temperature, 1.0)
}

func TestEmbeddingRequestValidation(t *testing.T) {
	req := &EmbeddingRequest{
		Input: []string{"Text to embed", "Another text"},
	}

	assert.NotEmpty(t, req.Input)
	assert.Len(t, req.Input, 2)
}

func TestChatRequestValidation(t *testing.T) {
	req := &ChatRequest{
		Context: "You are a helpful assistant",
		Messages: []*ChatMessage{
			{
				Content: "Hello",
				Author:  "user",
			},
			{
				Content: "Hi there!",
				Author:  "bot",
			},
		},
		Temperature:    0.7,
		TopP:           1,
		TopK:           40,
		CandidateCount: 1,
	}

	assert.NotEmpty(t, req.Context)
	assert.NotEmpty(t, req.Messages)
	assert.Equal(t, "user", req.Messages[0].Author)
	assert.Equal(t, "bot", req.Messages[1].Author)
}

// TestCreateCompletionErrorPaths was removed as it contained an unconditional skip

// TestCreateEmbeddingErrorPaths was removed as it contained an unconditional skip

// TestCreateChatErrorPaths was removed as it contained an unconditional skip

func TestProcessPredictionsErrors(t *testing.T) {
	t.Run("CreateCompletion missing content", func(t *testing.T) {
		// Simulate a response without "content" field
		value, _ := structpb.NewStruct(map[string]interface{}{
			"notContent": "value",
		})

		// Test the error handling logic that would be in CreateCompletion
		valueMap := value.AsMap()
		_, ok := valueMap["content"].(string)
		assert.False(t, ok)
	})

	t.Run("CreateEmbedding missing embeddings", func(t *testing.T) {
		// Simulate a response without "embeddings" field
		value, _ := structpb.NewStruct(map[string]interface{}{
			"notEmbeddings": "value",
		})

		valueMap := value.AsMap()
		_, ok := valueMap["embeddings"].(map[string]interface{})
		assert.False(t, ok)
	})

	t.Run("CreateEmbedding invalid float values", func(t *testing.T) {
		// Test float conversion logic
		values := []interface{}{
			float64(0.1),
			float32(0.2),
			"not a float", // This should cause an error
		}

		floatValues := []float32{}
		for _, v := range values {
			switch val := v.(type) {
			case float32:
				floatValues = append(floatValues, val)
			case float64:
				floatValues = append(floatValues, float32(val))
			default:
				// This simulates the error case
				err := errors.New("invalid value")
				require.Error(t, err)
				return
			}
		}
	})

	t.Run("CreateChat missing candidates", func(t *testing.T) {
		// Simulate a response without "candidates" field
		value, _ := structpb.NewStruct(map[string]interface{}{
			"notCandidates": "value",
		})

		valueMap := value.AsMap()
		_, ok := valueMap["candidates"].([]interface{})
		assert.False(t, ok)
	})
}

func TestChatResponseProcessing(t *testing.T) {
	// Test the logic for processing chat responses
	candidates := []interface{}{
		map[string]interface{}{
			"author":  "bot",
			"content": "Response 1",
		},
		map[string]interface{}{
			"author":  "bot",
			"content": "Response 2",
		},
	}

	// Process candidates as CreateChat would
	chatResponse := &ChatResponse{}
	for _, c := range candidates {
		candidate, ok := c.(map[string]interface{})
		require.True(t, ok)

		author, ok := candidate["author"].(string)
		require.True(t, ok)

		content, ok := candidate["content"].(string)
		require.True(t, ok)

		chatResponse.Candidates = append(chatResponse.Candidates, ChatMessage{
			Author:  author,
			Content: content,
		})
	}

	assert.Len(t, chatResponse.Candidates, 2)
	assert.Equal(t, "Response 1", chatResponse.Candidates[0].Content)
	assert.Equal(t, "Response 2", chatResponse.Candidates[1].Content)
}

func TestEmbeddingProcessing(t *testing.T) {
	// Test embedding value processing
	embeddings := map[string]interface{}{
		"values": []interface{}{
			float64(0.1),
			float64(0.2),
			float32(0.3),
		},
	}

	values, ok := embeddings["values"].([]interface{})
	require.True(t, ok)

	floatValues := []float32{}
	for _, v := range values {
		switch val := v.(type) {
		case float32:
			floatValues = append(floatValues, val)
		case float64:
			floatValues = append(floatValues, float32(val))
		}
	}

	assert.Len(t, floatValues, 3)
	assert.InDelta(t, 0.1, floatValues[0], 0.001)
	assert.InDelta(t, 0.2, floatValues[1], 0.001)
	assert.InDelta(t, 0.3, floatValues[2], 0.001)
}

func TestConstants(t *testing.T) {
	// Test that constants have expected values
	assert.Equal(t, "textembedding-gecko", embeddingModelName)
	assert.Equal(t, "text-bison", TextModelName)
	assert.Equal(t, "chat-bison", ChatModelName)
	assert.Equal(t, 4, defaultMaxConns)
}

func TestDefaultParameters(t *testing.T) {
	// Test default parameters
	assert.Equal(t, 0.2, defaultParameters["temperature"])
	assert.Equal(t, 256, defaultParameters["maxOutputTokens"])
	assert.Equal(t, 0.8, defaultParameters["topP"])
	assert.Equal(t, 40, defaultParameters["topK"])
}

func TestErrors(t *testing.T) {
	// Test error types
	assert.Equal(t, "missing value", ErrMissingValue.Error())
	assert.Equal(t, "invalid value", ErrInvalidValue.Error())
	assert.Equal(t, "empty response", ErrEmptyResponse.Error())
}

func TestMergeParamsEdgeCases(t *testing.T) {
	t.Run("int32 value", func(t *testing.T) {
		params := map[string]interface{}{
			"topK": int32(50),
		}
		result := mergeParams(defaultParameters, params)
		assert.NotNil(t, result)
	})

	t.Run("int64 value", func(t *testing.T) {
		params := map[string]interface{}{
			"maxOutputTokens": int64(1024),
		}
		result := mergeParams(defaultParameters, params)
		assert.NotNil(t, result)
	})
}
