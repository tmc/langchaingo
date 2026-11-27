// This file contains the partial schema of the Qdrant REST API.
// i.e. Only fields that are used by the application are specified.
// For a comprehensive reference of the Qdrant REST API
// Refer to https://qdrant.github.io/qdrant/redoc/

package qdrant

type upsertBatch struct {
	IDs      []string         `json:"ids"`
	Payloads []map[string]any `json:"payloads"`
	Vectors  [][]float32      `json:"vectors"`
}

type upsertBody struct {
	Batch upsertBatch `json:"batch"`
}

type result struct {
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload"`
}

type searchResponse struct {
	Result []result `json:"result"`
}

type namedVector struct {
	Vector []float32 `json:"vector"`
	Name   string    `json:"name"`
}

type searchBody struct {
	Vector         any     `json:"vector"`
	Filter         any     `json:"filter"`
	Limit          int     `json:"limit"`
	ScoreThreshold float32 `json:"score_threshold,omitempty"`
	WithVector     bool    `json:"with_vector"`
	WithPayload    bool    `json:"with_payload"`
}
