package pineconeClient

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Vector struct {
	Values   []float64         `json:"values"`
	Metadata map[string]string `json:"metadata"`
	ID       string            `json:"id"`
}

func NewVectorsFromValues(values [][]float64) []Vector {
	vectors := make([]Vector, 0)

	for i := 0; i < len(values); i++ {
		vectors = append(vectors, Vector{
			Values:   values[i],
			Metadata: make(map[string]string),
			ID:       uuid.New().String(),
		})
	}
	return vectors
}

type upsertPayload struct {
	Vectors   []Vector `json:"vectors"`
	Namespace string   `json:"namespace"`
}

type detail struct {
	TypeUrl string `json:"typeUrl"`
	Value   string `json:"value"`
}

type errorResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []detail `json:"details"`
}

func errorMessageFromErrorResponse(task string, body io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, body)
	if err != nil {
		return fmt.Errorf("error reading body of error message: %s", err.Error())
	}

	return fmt.Errorf("Error %s: body: %s", task, buf.String())
}

func (c Client) Upsert(ctx context.Context, vectors []Vector, nameSpace string) error {
	payload := upsertPayload{
		Vectors:   vectors,
		Namespace: nameSpace,
	}

	body, status, err := doRequest(ctx, payload, c.getEndpoint()+"/vectors/upsert", c.apiKey)
	if err != nil {
		return err
	}
	defer body.Close()

	if status == 200 {
		return nil
	}

	return errorMessageFromErrorResponse("upserting vectors", body)
}
