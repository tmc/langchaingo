package anthropic

import (
	"context"
	"time"

	"github.com/vxcontrol/langchaingo/llms/anthropic/internal/anthropicclient"
)

type ModelFeatures struct {
	Batch             bool
	Citations         bool
	CodeExecution     bool
	ContextManagement bool
	ImageInput        bool
	PDFInput          bool
	StructuredOutputs bool
}

// ModelInfo is one model from the vendor's listing. A false CapabilitiesReported
// means the record carried no capabilities object at all — read the other fields
// as unknown there, not as unsupported.
type ModelInfo struct {
	ID                   string
	DisplayName          string
	CreatedAt            time.Time
	MaxInputTokens       int
	MaxTokens            int
	CapabilitiesReported bool
	EffortSupported      bool
	EffortLevels         []string
	ThinkingSupported    bool
	ThinkingTypes        []string
	Features             ModelFeatures
}

// ListModelsRequest selects one page of the listing. Limit 0 takes the vendor's
// default page size.
type ListModelsRequest struct {
	Limit   int
	AfterID string
}

// ListModelsResponse is one page of the listing. HasMore true means further
// models exist beyond LastID; pass it as the next request's AfterID.
type ListModelsResponse struct {
	Models  []ModelInfo
	HasMore bool
	LastID  string
}

func (o *LLM) ListModels(ctx context.Context, req ListModelsRequest) (ListModelsResponse, error) {
	payload, err := o.client.ListModels(ctx, &anthropicclient.ModelsRequest{
		Limit:   req.Limit,
		AfterID: req.AfterID,
	})
	if err != nil {
		return ListModelsResponse{}, err
	}

	models := make([]ModelInfo, 0, len(payload.Data))
	for _, m := range payload.Data {
		models = append(models, modelInfoFrom(m))
	}
	return ListModelsResponse{Models: models, HasMore: payload.HasMore, LastID: payload.LastID}, nil
}

func modelInfoFrom(m anthropicclient.ModelInfo) ModelInfo {
	info := ModelInfo{
		ID:             m.ID,
		DisplayName:    m.DisplayName,
		CreatedAt:      m.CreatedAt,
		MaxInputTokens: m.MaxInputTokens,
		MaxTokens:      m.MaxTokens,
	}

	c := m.Capabilities
	if c == nil {
		return info
	}

	info.CapabilitiesReported = true
	info.EffortSupported = c.Effort.Supported
	info.ThinkingSupported = c.Thinking.Supported
	info.Features = ModelFeatures{
		Batch:             c.Batch.Supported,
		Citations:         c.Citations.Supported,
		CodeExecution:     c.CodeExecution.Supported,
		ContextManagement: c.ContextManagement.Supported,
		ImageInput:        c.ImageInput.Supported,
		PDFInput:          c.PDFInput.Supported,
		StructuredOutputs: c.StructuredOutputs.Supported,
	}

	levels := []struct {
		name string
		flag anthropicclient.ModelCapability
	}{
		{"low", c.Effort.Low}, {"medium", c.Effort.Medium}, {"high", c.Effort.High},
		{"xhigh", c.Effort.XHigh}, {"max", c.Effort.Max},
	}
	info.EffortLevels = make([]string, 0, len(levels))
	for _, lv := range levels {
		if lv.flag.Supported {
			info.EffortLevels = append(info.EffortLevels, lv.name)
		}
	}

	types := []struct {
		name string
		flag anthropicclient.ModelCapability
	}{
		{"enabled", c.Thinking.Types.Enabled}, {"adaptive", c.Thinking.Types.Adaptive},
	}
	info.ThinkingTypes = make([]string, 0, len(types))
	for _, ty := range types {
		if ty.flag.Supported {
			info.ThinkingTypes = append(info.ThinkingTypes, ty.name)
		}
	}
	return info
}
