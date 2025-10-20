package googleai

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestStructuredOutputIntegration(t *testing.T) {
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	llm, err := New(ctx,
		WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
		WithDefaultModel("gemini-2.0-flash"),
	)
	require.NoError(t, err)
	defer llm.Close()

	t.Run("simple user extraction", func(t *testing.T) {
		schema := &llms.StructuredOutputDefinition{
			Name:        "user_info",
			Description: "Extract user information",
			Schema: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"name": {
						Type:        llms.SchemaTypeString,
						Description: "User's full name",
					},
					"age": {
						Type:        llms.SchemaTypeInteger,
						Description: "User's age in years",
					},
					"email": {
						Type:        llms.SchemaTypeString,
						Description: "User's email address",
					},
				},
				Required: []string{"name", "age"},
			},
		}

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman,
				"Extract information about this user: John Doe is 30 years old and his email is john@example.com"),
		}

		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithStructuredOutput(schema),
			llms.WithMaxTokens(200),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		// Parse the JSON response
		var result struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Email string `json:"email"`
		}
		err = json.Unmarshal([]byte(resp.Choices[0].Content), &result)
		require.NoError(t, err)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, "john@example.com", result.Email)
	})

	t.Run("recipe extraction", func(t *testing.T) {
		schema := &llms.StructuredOutputDefinition{
			Name:        "recipe",
			Description: "Extract recipe information",
			Schema: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"title": {
						Type:        llms.SchemaTypeString,
						Description: "Recipe title",
					},
					"ingredients": {
						Type:        llms.SchemaTypeArray,
						Description: "List of ingredients",
						Items: &llms.StructuredOutputSchema{
							Type: llms.SchemaTypeString,
						},
					},
					"steps": {
						Type:        llms.SchemaTypeArray,
						Description: "Cooking steps",
						Items: &llms.StructuredOutputSchema{
							Type: llms.SchemaTypeString,
						},
					},
				},
				Required: []string{"title", "ingredients", "steps"},
			},
		}

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman,
				"Give me a simple recipe for scrambled eggs"),
		}

		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithStructuredOutput(schema),
			llms.WithMaxTokens(300),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		// Parse the JSON response
		var result struct {
			Title       string   `json:"title"`
			Ingredients []string `json:"ingredients"`
			Steps       []string `json:"steps"`
		}
		err = json.Unmarshal([]byte(resp.Choices[0].Content), &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result.Title)
		assert.NotEmpty(t, result.Ingredients)
		assert.NotEmpty(t, result.Steps)
		t.Logf("Recipe: %s", result.Title)
		t.Logf("Ingredients: %v", result.Ingredients)
		t.Logf("Steps: %v", result.Steps)
	})

	t.Run("nested object", func(t *testing.T) {
		schema := &llms.StructuredOutputDefinition{
			Name: "company_info",
			Schema: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"company_name": {Type: llms.SchemaTypeString},
					"ceo": {
						Type: llms.SchemaTypeObject,
						Properties: map[string]*llms.StructuredOutputSchema{
							"name": {Type: llms.SchemaTypeString},
							"age":  {Type: llms.SchemaTypeInteger},
						},
						Required: []string{"name"},
					},
				},
				Required: []string{"company_name", "ceo"},
			},
		}

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman,
				"Extract: Anthropic is led by Dario Amodei, who is 41 years old"),
		}

		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithStructuredOutput(schema),
			llms.WithMaxTokens(150),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		var result struct {
			CompanyName string `json:"company_name"`
			CEO         struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			} `json:"ceo"`
		}
		err = json.Unmarshal([]byte(resp.Choices[0].Content), &result)
		require.NoError(t, err)

		assert.Contains(t, result.CompanyName, "Anthropic")
		assert.Contains(t, result.CEO.Name, "Dario")
	})

	t.Run("enum values", func(t *testing.T) {
		schema := &llms.StructuredOutputDefinition{
			Name: "sentiment",
			Schema: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"sentiment": {
						Type:        llms.SchemaTypeString,
						Description: "Sentiment of the text",
						Enum:        []string{"positive", "negative", "neutral"},
					},
					"confidence": {
						Type:        llms.SchemaTypeNumber,
						Description: "Confidence score 0-1",
					},
				},
				Required: []string{"sentiment"},
			},
		}

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman,
				"Analyze sentiment: This product is amazing! I love it!"),
		}

		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithStructuredOutput(schema),
			llms.WithMaxTokens(100),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		var result struct {
			Sentiment  string  `json:"sentiment"`
			Confidence float64 `json:"confidence"`
		}
		err = json.Unmarshal([]byte(resp.Choices[0].Content), &result)
		require.NoError(t, err)

		assert.Equal(t, "positive", result.Sentiment)
		t.Logf("Sentiment: %s, Confidence: %.2f", result.Sentiment, result.Confidence)
	})
}
