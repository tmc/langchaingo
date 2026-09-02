package anthropicclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ModelCapability struct {
	Supported bool `json:"supported"`
}

type ModelEffortCapability struct {
	Supported bool            `json:"supported"`
	Low       ModelCapability `json:"low"`
	Medium    ModelCapability `json:"medium"`
	High      ModelCapability `json:"high"`
	XHigh     ModelCapability `json:"xhigh"`
	Max       ModelCapability `json:"max"`
}

type ModelThinkingTypes struct {
	Enabled  ModelCapability `json:"enabled"`
	Adaptive ModelCapability `json:"adaptive"`
}

type ModelThinkingCapability struct {
	Supported bool               `json:"supported"`
	Types     ModelThinkingTypes `json:"types"`
}

type ModelCapabilities struct {
	Batch             ModelCapability         `json:"batch"`
	Citations         ModelCapability         `json:"citations"`
	CodeExecution     ModelCapability         `json:"code_execution"`
	ContextManagement ModelCapability         `json:"context_management"`
	Effort            ModelEffortCapability   `json:"effort"`
	ImageInput        ModelCapability         `json:"image_input"`
	PDFInput          ModelCapability         `json:"pdf_input"`
	StructuredOutputs ModelCapability         `json:"structured_outputs"`
	Thinking          ModelThinkingCapability `json:"thinking"`
}

type ModelInfo struct {
	Type           string             `json:"type"`
	ID             string             `json:"id"`
	DisplayName    string             `json:"display_name"`
	CreatedAt      time.Time          `json:"created_at"`
	MaxInputTokens int                `json:"max_input_tokens"`
	MaxTokens      int                `json:"max_tokens"`
	Capabilities   *ModelCapabilities `json:"capabilities"`
}

type ModelsRequest struct {
	Limit   int
	AfterID string
}

type ModelsResponsePayload struct {
	Data    []ModelInfo `json:"data"`
	HasMore bool        `json:"has_more"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
}

func (c *Client) ListModels(ctx context.Context, r *ModelsRequest) (*ModelsResponsePayload, error) {
	path := "/models"
	if r != nil {
		q := url.Values{}
		if r.Limit > 0 {
			q.Set("limit", strconv.Itoa(r.Limit))
		}
		if r.AfterID != "" {
			q.Set("after_id", r.AfterID)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}

	resp, err := c.request(ctx, http.MethodGet, path, http.NoBody, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	var payload ModelsResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode models response: %w", err)
	}
	return &payload, nil
}
