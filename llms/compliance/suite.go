package compliance

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/0xDezzy/langchaingo/llms"
)

// Suite tests provider compliance with the LLM interface.
type Suite struct {
	// Provider is the provider name.
	Provider string

	// Model is the model to test.
	Model llms.Model

	// SkipTests contains test names to skip.
	SkipTests map[string]bool

	// Timeout for individual tests.
	Timeout time.Duration
}

// NewSuite creates a new compliance test suite.
func NewSuite(provider string, model llms.Model) *Suite {
	return &Suite{
		Provider:  provider,
		Model:     model,
		SkipTests: make(map[string]bool),
		Timeout:   30 * time.Second,
	}
}

// Skip marks a test to be skipped.
func (s *Suite) Skip(testName string) {
	s.SkipTests[testName] = true
}

// Run executes all compliance tests.
func (s *Suite) Run(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"BasicGeneration", s.testBasicGeneration},
		{"MultiMessage", s.testMultiMessage},
		{"Temperature", s.testTemperature},
		{"MaxTokens", s.testMaxTokens},
		{"StopSequences", s.testStopSequences},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if s.SkipTests[test.name] {
				t.Skip("Test skipped by configuration")
			}
			test.fn(t)
		})
	}
}

func (s *Suite) testBasicGeneration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	content := []llms.MessageContent{
		{Role: "user", Parts: []llms.ContentPart{llms.TextPart("Say 'Hello, World!' and nothing else.")}},
	}

	resp, err := s.Model.GenerateContent(ctx, content)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices returned")
	}

	output := resp.Choices[0].Content
	if !strings.Contains(strings.ToLower(output), "hello") || !strings.Contains(strings.ToLower(output), "world") {
		t.Errorf("Expected 'Hello, World!' but got: %s", output)
	}
}

func (s *Suite) testMultiMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	content := []llms.MessageContent{
		{Role: "user", Parts: []llms.ContentPart{llms.TextPart("My name is Alice.")}},
		{Role: "assistant", Parts: []llms.ContentPart{llms.TextPart("Nice to meet you, Alice!")}},
		{Role: "user", Parts: []llms.ContentPart{llms.TextPart("What's my name?")}},
	}

	resp, err := s.Model.GenerateContent(ctx, content)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices returned")
	}

	output := resp.Choices[0].Content
	if !strings.Contains(strings.ToLower(output), "alice") {
		t.Errorf("Expected response to mention 'Alice' but got: %s", output)
	}
}

func (s *Suite) testTemperature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	content := []llms.MessageContent{
		{Role: "user", Parts: []llms.ContentPart{llms.TextPart("Write the number 42.")}},
	}

	// Test with temperature=0 (deterministic)
	resp1, err := s.Model.GenerateContent(ctx, content, llms.WithTemperature(0.0))
	if err != nil {
		t.Fatalf("GenerateContent with temperature=0 failed: %v", err)
	}

	// Test with temperature=1 (creative)
	resp2, err := s.Model.GenerateContent(ctx, content, llms.WithTemperature(1.0))
	if err != nil {
		t.Fatalf("GenerateContent with temperature=1 failed: %v", err)
	}

	if len(resp1.Choices) == 0 || len(resp2.Choices) == 0 {
		t.Fatal("No choices returned")
	}

	// Both should contain "42"
	output1 := resp1.Choices[0].Content
	output2 := resp2.Choices[0].Content

	if !strings.Contains(output1, "42") {
		t.Errorf("Expected '42' in temperature=0 output but got: %s", output1)
	}
	if !strings.Contains(output2, "42") {
		t.Errorf("Expected '42' in temperature=1 output but got: %s", output2)
	}
}

func (s *Suite) testMaxTokens(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	content := []llms.MessageContent{
		{Role: "user", Parts: []llms.ContentPart{llms.TextPart("Count from 1 to 100.")}},
	}

	// Request a very short response
	resp, err := s.Model.GenerateContent(ctx, content, llms.WithMaxTokens(10))
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices returned")
	}

	output := resp.Choices[0].Content
	// Very rough check - with 10 tokens, we shouldn't get past "10" or so
	if strings.Contains(output, "50") || strings.Contains(output, "100") {
		t.Errorf("Expected short response with max_tokens=10, but got: %s", output)
	}
}

func (s *Suite) testStopSequences(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	content := []llms.MessageContent{
		{Role: "user", Parts: []llms.ContentPart{llms.TextPart("List the days of the week.")}},
	}

	// Stop at "Wednesday"
	resp, err := s.Model.GenerateContent(ctx, content, llms.WithStopWords([]string{"Wednesday"}))
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices returned")
	}

	output := resp.Choices[0].Content
	// Should contain Monday and Tuesday but not Thursday
	lowerOutput := strings.ToLower(output)
	if !strings.Contains(lowerOutput, "monday") || !strings.Contains(lowerOutput, "tuesday") {
		t.Errorf("Expected Monday and Tuesday in output but got: %s", output)
	}
	if strings.Contains(lowerOutput, "thursday") || strings.Contains(lowerOutput, "friday") {
		t.Errorf("Expected output to stop before Thursday but got: %s", output)
	}
}
