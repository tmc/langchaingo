package bedrock

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const (
	/*
		ModelCohereEn is the model id for the cohere english embeddings.

		  ModelDimensions := 1024
		  MaxTokens := 512
		  Languages := []string{"English"}
	*/
	ModelCohereEn = "cohere.embed-english-v3"

	/*
		ModelCohereMulti is the model id for the cohere multilingual embeddings.

		  ModelDimensions := 1024
		  MaxTokens:= 512
		  Languages := [108]string
	*/
	ModelCohereMulti = "cohere.embed-multilingual-v3"
)

const (
	// CohereInputTypeText is the input type for text embeddings.
	CohereInputTypeText = "search_document"
	// CohereInputTypeQuery is the input type for query embeddings.
	CohereInputTypeQuery = "search_query"
)

type cohereTextEmbeddingsInput struct {
	Texts     []string `json:"texts"`
	InputType string   `json:"input_type"`
}

type cohereTextEmbeddingsOutput struct {
	ResponseType string      `json:"response_type"`
	Embeddings   [][]float32 `json:"embeddings"`
}

func FetchCohereTextEmbeddings(
	ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	inputs []string,
	inputType string,
) ([][]float32, error) {
	var err error

	bodyStruct := cohereTextEmbeddingsInput{
		Texts:     inputs,
		InputType: inputType,
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
	var response cohereTextEmbeddingsOutput
	err = json.Unmarshal(result.Body, &response)
	if err != nil {
		return nil, err
	}

	return response.Embeddings, nil
}
