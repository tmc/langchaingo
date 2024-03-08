package bedrock

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const (
	/*
		ModelTitanEmbedG1 is the model id for the amazon text embeddings.

		  MaxTokens := 8000
		  ModelDimensions := 1536
		  Languages := []string{"English", "Arabic", "Chinese (Simplified)", "French", "German", "Hindi", "Japanese", "Spanish", "Czech", "Filipino", "Hebrew", "Italian", "Korean", "Portuguese", "Russian", "Swedish", "Turkish", "Chinese (Traditional)", "Dutch", "Kannada", "Malayalam", "Marathi", "Polish", "Tamil", "Telugu", ...}
	*/
	ModelTitanEmbedG1 = "amazon.titan-embed-text-v1"
)

type amazonEmbeddingsInput struct {
	InputText string `json:"inputText"`
}

type amazonEmbeddingsOutput struct {
	Embedding []float32 `json:"embedding"`
}

func FetchAmazonTextEmbeddings(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	texts []string,
) ([][]float32, error) {
	embeddings := make([][]float32, 0, len(texts))

	for _, text := range texts {
		bodyStruct := amazonEmbeddingsInput{
			InputText: text,
		}
		body, err := json.Marshal(bodyStruct)
		if err != nil {
			return nil, err
		}
		modelInput := &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(modelID),
			Accept:      aws.String("*/*"),
			ContentType: aws.String("application/json"),
			Body:        body,
		}

		result, err := client.InvokeModel(ctx, modelInput)
		if err != nil {
			return nil, err
		}

		var response amazonEmbeddingsOutput
		err = json.Unmarshal(result.Body, &response)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, response.Embedding)
	}

	return embeddings, nil
}
