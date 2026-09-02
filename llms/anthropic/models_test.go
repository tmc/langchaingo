package anthropic_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms/anthropic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listingCall struct {
	method   string
	path     string
	rawQuery string
}

func serveListing(t *testing.T, status int, body string) (*anthropic.LLM, *listingCall) {
	t.Helper()

	seen := new(listingCall)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.method, seen.path, seen.rawQuery = r.Method, r.URL.Path, r.URL.RawQuery
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("test"), anthropic.WithBaseURL(srv.URL))
	require.NoError(t, err)
	return llm, seen
}

const vendorSonnet46 = `{"data":[{"type":"model","id":"claude-sonnet-4-6",` +
	`"display_name":"Claude Sonnet 4.6","created_at":"2026-02-04T00:00:00Z",` +
	`"max_input_tokens":1000000,"max_tokens":128000,"capabilities":{` +
	`"batch":{"supported":true},"citations":{"supported":true},` +
	`"code_execution":{"supported":true},"context_management":{"supported":true},` +
	`"image_input":{"supported":true},"pdf_input":{"supported":true},` +
	`"structured_outputs":{"supported":true},` +
	`"effort":{"supported":true,"low":{"supported":true},"medium":{"supported":true},` +
	`"high":{"supported":true},"xhigh":{"supported":false},"max":{"supported":true}},` +
	`"thinking":{"supported":true,"types":{"enabled":{"supported":true},` +
	`"adaptive":{"supported":true}}}}}],"has_more":false,"first_id":"claude-sonnet-4-6",` +
	`"last_id":"claude-sonnet-4-6"}`

func TestTheListingCarriesEveryCapabilityTheVendorReports(t *testing.T) {
	t.Parallel()

	llm, _ := serveListing(t, http.StatusOK, vendorSonnet46)

	resp, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Models, 1)

	m := resp.Models[0]
	assert.Equal(t, "claude-sonnet-4-6", m.ID)
	assert.Equal(t, "Claude Sonnet 4.6", m.DisplayName)
	assert.Equal(t, 2026, m.CreatedAt.Year())
	assert.Equal(t, 1000000, m.MaxInputTokens)
	assert.Equal(t, 128000, m.MaxTokens)
	assert.True(t, m.CapabilitiesReported)
	assert.True(t, m.EffortSupported)
	assert.Equal(t, []string{"low", "medium", "high", "max"}, m.EffortLevels)
	assert.True(t, m.ThinkingSupported)
	assert.Equal(t, []string{"enabled", "adaptive"}, m.ThinkingTypes)
	assert.True(t, m.Features.StructuredOutputs)
	assert.True(t, m.Features.ContextManagement)
	assert.False(t, resp.HasMore)
	assert.Equal(t, "claude-sonnet-4-6", resp.LastID)
}

func TestAModelOffTheEffortScaleReportsNoLevels(t *testing.T) {
	t.Parallel()

	const body = `{"data":[{"type":"model","id":"claude-haiku-4-5-20251001",` +
		`"max_input_tokens":200000,"max_tokens":64000,"capabilities":{` +
		`"effort":{"supported":false,"low":{"supported":false},"medium":{"supported":false},` +
		`"high":{"supported":false},"xhigh":{"supported":false},"max":{"supported":false}},` +
		`"thinking":{"supported":true,"types":{"enabled":{"supported":true},` +
		`"adaptive":{"supported":false}}}}}],"has_more":false}`

	llm, _ := serveListing(t, http.StatusOK, body)

	resp, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Models, 1)

	m := resp.Models[0]
	assert.True(t, m.CapabilitiesReported)
	assert.False(t, m.EffortSupported)
	assert.Empty(t, m.EffortLevels)
	assert.True(t, m.ThinkingSupported)
	assert.Equal(t, []string{"enabled"}, m.ThinkingTypes)
}

func TestARecordWithoutCapabilitiesIsNotReadAsUnsupported(t *testing.T) {
	t.Parallel()

	const body = `{"data":[{"type":"model","id":"claude-newborn-9",` +
		`"display_name":"Claude Newborn 9"}],"has_more":false}`

	llm, _ := serveListing(t, http.StatusOK, body)

	resp, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Models, 1)

	m := resp.Models[0]
	assert.Equal(t, "claude-newborn-9", m.ID)
	assert.False(t, m.CapabilitiesReported,
		"the vendor said nothing about this model, which is not the same as saying nothing is supported")
	assert.Empty(t, m.EffortLevels)
	assert.Empty(t, m.ThinkingTypes)
}

func TestTheListingIsFetchedWithAGetAndTheRequestedPage(t *testing.T) {
	t.Parallel()

	llm, seen := serveListing(t, http.StatusOK, vendorSonnet46)

	_, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{
		Limit:   1000,
		AfterID: "claude-opus-5",
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, seen.method)
	assert.Equal(t, "/models", seen.path)
	assert.Equal(t, "after_id=claude-opus-5&limit=1000", seen.rawQuery)
}

func TestAnUnpagedListingSendsNoQuery(t *testing.T) {
	t.Parallel()

	llm, seen := serveListing(t, http.StatusOK, vendorSonnet46)

	_, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{})
	require.NoError(t, err)
	assert.Empty(t, seen.rawQuery)
}

func TestARefusedListingSurfacesTheVendorMessage(t *testing.T) {
	t.Parallel()

	const body = `{"type":"error","error":{"type":"authentication_error",` +
		`"message":"invalid x-api-key"}}`

	llm, _ := serveListing(t, http.StatusUnauthorized, body)

	_, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid x-api-key")
	assert.Contains(t, err.Error(), "401")
}
