package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// Define some tools for the model to use
var (
	calculateTool = llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "calculate",
			Description: "Perform mathematical calculations",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "Mathematical expression to evaluate",
					},
				},
				"required": []string{"expression"},
			},
		},
	}

	searchTool = llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "search_knowledge",
			Description: "Search for information in a knowledge base",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	analyzeTool = llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "analyze_data",
			Description: "Analyze data and return insights",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type":        "array",
						"description": "Data points to analyze",
						"items": map[string]interface{}{
							"type": "number",
						},
					},
					"analysis_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of analysis: mean, median, std_dev, trend",
					},
				},
				"required": []string{"data", "analysis_type"},
			},
		},
	}

	// New tool that depends on results from other tools
	generateReportTool = llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "generate_report",
			Description: "Generate a strategic report based on analysis results. ONLY use after completing all calculations and analyses.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"growth_rate": map[string]interface{}{
						"type":        "number",
						"description": "Year-over-year growth rate from calculations",
					},
					"trend_direction": map[string]interface{}{
						"type":        "string",
						"description": "Trend direction from analysis (upward/downward/stable)",
					},
					"volatility_level": map[string]interface{}{
						"type":        "string",
						"description": "Volatility level based on std deviation (low/moderate/high)",
					},
					"key_insights": map[string]interface{}{
						"type":        "array",
						"description": "Key insights from research",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"prediction": map[string]interface{}{
						"type":        "number",
						"description": "Q1-2025 predicted value",
					},
				},
				"required": []string{"growth_rate", "trend_direction", "volatility_level", "key_insights", "prediction"},
			},
		},
	}

	// Tool for making final predictions based on all data
	makePredictionTool = llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "make_prediction",
			Description: "Make predictions based on calculated metrics. Requires growth rates and averages from previous calculations.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"base_value": map[string]interface{}{
						"type":        "number",
						"description": "The base value to predict from (e.g., Q4-2024 value)",
					},
					"growth_rate": map[string]interface{}{
						"type":        "number",
						"description": "Growth rate to apply (as decimal, e.g., 0.05 for 5%)",
					},
					"seasonality_factor": map[string]interface{}{
						"type":        "number",
						"description": "Seasonal adjustment factor",
					},
					"confidence_level": map[string]interface{}{
						"type":        "string",
						"description": "Confidence level: high, medium, low",
					},
				},
				"required": []string{"base_value", "growth_rate", "seasonality_factor", "confidence_level"},
			},
		},
	}
)

// debugTransport wraps http.RoundTripper to log requests and responses
type debugTransport struct {
	Transport http.RoundTripper
	Debug     bool
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.Debug {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("HTTP REQUEST")
		fmt.Println(strings.Repeat("-", 80))

		// Dump request
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Printf("Error dumping request: %v\n", err)
		} else {
			fmt.Printf("%s\n", reqDump)
		}
	}

	// Make the actual request
	resp, err := d.Transport.RoundTrip(req)

	if d.Debug && resp != nil {
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("HTTP RESPONSE")
		fmt.Println(strings.Repeat("-", 80))

		// Dump response
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Printf("Error dumping response: %v\n", err)
		} else {
			fmt.Printf("%s\n", respDump)
		}
		fmt.Println(strings.Repeat("=", 80) + "\n")
	}

	return resp, err
}

func main() {
	// Parse command line flags
	debugHTTP := flag.Bool("debug-http", false, "Show raw HTTP requests and responses")
	flag.Parse()

	ctx := context.Background()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║      Claude 4 Interleaved Thinking Demo                   ║")
	fmt.Println("║     Demonstrating thinking between tool calls             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if *debugHTTP {
		fmt.Println("🔍 DEBUG MODE: HTTP requests/responses will be displayed")
		fmt.Println()
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY not set. Skipping demo.")
		fmt.Println("   Set the environment variable to run this example:")
		fmt.Println("   export ANTHROPIC_API_KEY=your-api-key")
		return
	}

	// Stage 1: Initialization
	fmt.Println("🚀 STAGE 1: INITIALIZATION")
	fmt.Println("─────────────────────────")
	fmt.Println("  • Model: Claude Sonnet 4.5")
	fmt.Println("  • Beta: interleaved-thinking-2025-05-14")
	fmt.Println("  • Feature: Thinking between tool calls")

	// Configure options for Anthropic client
	anthropicOpts := []anthropic.Option{
		anthropic.WithModel("claude-sonnet-4-5"),
		// IMPORTANT: Set cache strategy at CLIENT level for automatic application to ALL requests
		// This ensures caching works correctly across multi-turn conversations
		anthropic.WithDefaultCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true, // Cache tool definitions (saves ~90% on tools after first request)
			CacheSystem:   true, // Cache system prompt (saves ~90% on system prompt after first request)
			CacheMessages: true, // Cache conversation history (incremental writes, 90% read savings)
			TTL:           "1h", // 1-hour cache for extended thinking sessions
		}),
	}

	// Add debug HTTP client if flag is set
	if *debugHTTP {
		httpClient := &http.Client{
			Transport: &debugTransport{
				Transport: http.DefaultTransport,
				Debug:     true,
			},
		}
		anthropicOpts = append(anthropicOpts, anthropic.WithHTTPClient(httpClient))
	}

	// Using Claude Sonnet 4 for interleaved thinking
	llm, err := anthropic.New(anthropicOpts...)
	if err != nil {
		fmt.Printf("  ❌ Error initializing Anthropic: %v\n", err)
		return
	}
	fmt.Println("  ✅ Model initialized successfully")
	fmt.Println()

	// Stage 2: Problem Setup
	fmt.Println("📊 STAGE 2: PROBLEM SETUP")
	fmt.Println("─────────────────────────")
	fmt.Println("  Complex multi-step analysis task:")
	fmt.Println("  • Quarterly sales data analysis")
	fmt.Println("  • Year-over-year growth calculation")
	fmt.Println("  • Trend analysis and prediction")
	fmt.Println()

	// Add unique marker to avoid cache collisions from previous runs
	runID := fmt.Sprintf("Run-%d", time.Now().Unix())

	// Complex multi-step problem requiring tool use and reasoning
	prompt := fmt.Sprintf(`[%s] You're helping a data scientist analyze quarterly sales data and make strategic decisions.`, runID) + `

The quarterly sales (in millions) for the last 8 quarters are:
Q1-2023: 12.5, Q2-2023: 14.2, Q3-2023: 13.8, Q4-2023: 16.1
Q1-2024: 15.3, Q2-2024: 17.4, Q3-2024: 16.9, Q4-2024: 19.2

IMPORTANT: Complete this multi-stage analysis. Some tools depend on results from others, creating natural stages:

STAGE 1 - PARALLEL DATA GATHERING (invoke all these tools simultaneously):
1. CALCULATIONS:
   - Year-over-year growth rate for Q4: (19.2 - 16.1) / 16.1 * 100
   - Average quarterly growth 2024: ((17.4/15.3 - 1) + (16.9/17.4 - 1) + (19.2/16.9 - 1)) / 3
   - Overall growth Q1-2023 to Q4-2024: (19.2 - 12.5) / 12.5 * 100
   - Average 2023: (12.5 + 14.2 + 13.8 + 16.1) / 4
   - Average 2024: (15.3 + 17.4 + 16.9 + 19.2) / 4

2. DATA ANALYSIS:
   - Trend analysis on [12.5, 14.2, 13.8, 16.1, 15.3, 17.4, 16.9, 19.2]
   - Standard deviation of the same data
   - Mean of all data points

3. RESEARCH:
   - "seasonal sales patterns in retail"
   - "factors driving quarterly sales growth"
   - "economic indicators affecting sales performance"

STAGE 2 - SYNTHESIS (use make_prediction tool AFTER Stage 1 completes):
After receiving all Stage 1 results, use the make_prediction tool with:
- base_value: Q4-2024 value (19.2)
- growth_rate: Use the average quarterly growth rate from Stage 1
- seasonality_factor: Derive from the seasonal patterns research
- confidence_level: Based on the standard deviation analysis

STAGE 3 - REPORT GENERATION (use generate_report tool AFTER Stage 2):
Finally, use generate_report to create a comprehensive summary using ALL previous results.

This demonstrates interleaved thinking: parallel execution where possible, sequential when dependencies exist.`

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(prompt)},
		},
	}

	// Stage 3: Configuration
	fmt.Println("⚙️  STAGE 3: CONFIGURATION")
	fmt.Println("─────────────────────────")
	fmt.Println("  • Thinking Mode: MEDIUM")
	fmt.Println("  • Interleaved Thinking: ENABLED")
	fmt.Println("  • Temperature: 1.0 (required for thinking)")
	fmt.Println("  • Max Tokens: 4000")
	fmt.Println("  • Prompt Caching: ENABLED (1-hour TTL)")
	fmt.Println("  • Cache Strategy:")
	fmt.Println("    - Tools: CACHED (5 tools)")
	fmt.Println("    - System: NOT USED (no system prompt)")
	fmt.Println("    - Messages: CACHED (conversation history)")
	fmt.Println("  • Tools Available:")
	fmt.Println("    - calculate: Mathematical calculations")
	fmt.Println("    - search_knowledge: Information retrieval")
	fmt.Println("    - analyze_data: Statistical analysis")
	fmt.Println()

	// Cache analytics tracking
	type CacheStats struct {
		Iteration        int
		CacheCreation    int
		CacheRead        int
		InputTokens      int
		OutputTokens     int
		Cache5mCreation  int
		Cache1hCreation  int
		TotalTokens      int
		EffectiveCost    float64 // Relative cost considering cache
		SavingsVsNoCache float64 // Savings percentage
	}
	var cacheStatsList []CacheStats

	// Configure with interleaved thinking for tool use
	var streamedContent strings.Builder
	var contentBlockCount int

	opts := []llms.CallOption{
		// Enable thinking mode for reasoning between tools
		llms.WithReasoning(llms.ReasoningMedium, 4000),
		// Add interleaved thinking beta header
		anthropic.WithInterleavedThinking(),
		// Enable prompt caching (CRITICAL for interleaved thinking with tools!)
		anthropic.WithPromptCaching(),
		// Note: Cache strategy is set at CLIENT level (see anthropic.New above)
		// This ensures it applies to ALL requests automatically
		// Temperature must be 1 when thinking is enabled
		llms.WithTemperature(1.0),
		// Provide tools
		llms.WithTools([]llms.Tool{
			calculateTool,
			searchTool,
			analyzeTool,
			makePredictionTool,
			generateReportTool,
		}),
		llms.WithMaxTokens(8000), // Increased for multiple tool calls
		// Add streaming to show progress
		// Note: The streaming function receives processed text content,
		// not raw SSE events. The anthropic client handles SSE parsing internally.
		llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			chunkStr := chunk.Content

			// Show progress dots for content generation
			if len(chunkStr) > 0 {
				if contentBlockCount == 0 {
					contentBlockCount++
					fmt.Printf("\n\n  📝 GENERATING RESPONSE: ")
					fmt.Print("\n  │  ")
				}
				fmt.Print("•")
				streamedContent.WriteString(chunkStr)
			}
			return nil
		}),
	}

	// Stage 4: Processing
	fmt.Println("🔄 STAGE 4: PROCESSING")
	fmt.Println("─────────────────────────")
	fmt.Println("  Starting multi-step analysis with interleaved thinking...")

	// Helper function to collect cache statistics
	collectCacheStats := func(resp *llms.ContentResponse, iterNum int) CacheStats {
		genInfo := resp.Choices[0].GenerationInfo
		stats := CacheStats{Iteration: iterNum}

		if v, ok := genInfo["CacheCreationInputTokens"].(int); ok {
			stats.CacheCreation = v
		}
		if v, ok := genInfo["CacheReadInputTokens"].(int); ok {
			stats.CacheRead = v
		}
		if v, ok := genInfo["PromptTokens"].(int); ok {
			stats.InputTokens = v
		}
		if v, ok := genInfo["OutputTokens"].(int); ok {
			stats.OutputTokens = v
		}
		if v, ok := genInfo["CacheCreationEphemeral5mInputTokens"].(int); ok {
			stats.Cache5mCreation = v
		}
		if v, ok := genInfo["CacheCreationEphemeral1hInputTokens"].(int); ok {
			stats.Cache1hCreation = v
		}
		if v, ok := genInfo["TotalTokens"].(int); ok {
			stats.TotalTokens = v
		} else {
			stats.TotalTokens = stats.InputTokens + stats.CacheRead + stats.OutputTokens
		}

		// Calculate effective cost (relative to no-cache baseline)
		// Cache write: 125% cost, Cache read: 10% cost, Regular input: 100% cost
		regularInput := stats.InputTokens - stats.CacheCreation
		stats.EffectiveCost = float64(regularInput)*1.0 +
			float64(stats.CacheCreation)*1.25 +
			float64(stats.CacheRead)*0.1 +
			float64(stats.OutputTokens)*1.0

		// Calculate savings vs no-cache scenario
		noCacheCost := float64(stats.TotalTokens) * 1.0
		if noCacheCost > 0 {
			stats.SavingsVsNoCache = ((noCacheCost - stats.EffectiveCost) / noCacheCost) * 100
		}

		return stats
	}

	// Make initial request
	resp, err := llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		fmt.Printf("\n  ❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n  ℹ️  Initial response received (streaming may still be in progress)\n")
	time.Sleep(10 * time.Second)

	// Collect cache stats for initial request
	if resp != nil && len(resp.Choices) > 0 {
		stats := collectCacheStats(resp, 0)
		cacheStatsList = append(cacheStatsList, stats)
	}

	// Handle tool calls in a conversation loop
	maxIterations := 10 // Prevent infinite loops
	iteration := 0
	allResponses := []llms.ContentChoice{} // Store all responses for final display

	// Store initial response
	if resp != nil && len(resp.Choices) > 0 {
		allResponses = append(allResponses, *resp.Choices[0])
	}

	for iteration < maxIterations && resp != nil && len(resp.Choices) > 0 {
		choice := resp.Choices[0]

		// Check if there are tool calls to process
		// Note: Tool calls appear in the response AFTER streaming completes
		if len(choice.ToolCalls) > 0 {
			fmt.Printf("\n\n  🔄 Iteration %d: Processing %d tool calls...\n", iteration+1, len(choice.ToolCalls))
			for i, tc := range choice.ToolCalls {
				fmt.Printf("      Tool %d: %s\n", i+1, tc.FunctionCall.Name)
			}

			// Add assistant message with tool calls to conversation
			// IMPORTANT: Must include thinking blocks (reasoning) for interleaved thinking to work
			// According to Anthropic docs: "When using extended thinking with tool use,
			// you must pass thinking blocks back to the API"
			assistantParts := []llms.ContentPart{
				llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
			}
			assistantParts = append(assistantParts, convertToolCallsToParts(choice.ToolCalls)...)

			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: assistantParts,
			})

			// Execute tools and add results to conversation
			for _, tc := range choice.ToolCalls {
				result := executeToolCall(tc.FunctionCall.Name, tc.FunctionCall.Arguments)

				// Add tool result to messages
				messages = append(messages, llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: tc.ID,
							Name:       tc.FunctionCall.Name,
							Content:    result,
						},
					},
				})
			}

			// Continue conversation with tool results
			fmt.Printf("\n  📤 Sending tool results back to model (iteration %d)...\n", iteration+1)

			// Remove streaming for subsequent calls to avoid duplicate output
			continuationOpts := make([]llms.CallOption, 0, len(opts))
			for _, opt := range opts {
				continuationOpts = append(continuationOpts, opt)
			}
			// Override streaming for continuation
			continuationOpts = append(continuationOpts, llms.WithStreamingFunc(nil))

			resp, err = llm.GenerateContent(ctx, messages, continuationOpts...)
			if err != nil {
				fmt.Printf("\n  ⚠️  Error in iteration %d: %v\n", iteration+1, err)
				break
			}

			// Store this response and collect cache stats
			if resp != nil && len(resp.Choices) > 0 {
				allResponses = append(allResponses, *resp.Choices[0])
				fmt.Printf("\n  ✅ Received response with %d tool calls\n", len(resp.Choices[0].ToolCalls))

				// Collect cache statistics for this iteration
				stats := collectCacheStats(resp, iteration+1)
				cacheStatsList = append(cacheStatsList, stats)
			}

			iteration++
		} else {
			// No more tool calls, conversation complete
			break
		}
	}

	if iteration >= maxIterations {
		fmt.Println("\n  ⚠️  Reached maximum iterations, stopping conversation loop")
	}

	// Ensure we have a clean line after streaming
	fmt.Println("─────────────────────────")

	// Stage 5: Results Analysis
	fmt.Println("✅ STAGE 5: RESULTS ANALYSIS")
	fmt.Println("─────────────────────────")

	// Show final iteration count and tool call summary
	fmt.Printf("\n  📊 Completed after %d iteration(s)\n", iteration)

	// Count total tool calls processed
	totalToolCalls := 0
	if resp != nil && len(resp.Choices) > 0 {
		totalToolCalls = len(resp.Choices[0].ToolCalls)
		if totalToolCalls > 0 {
			fmt.Printf("  ✅ Total tool calls executed: %d\n", totalToolCalls)

			// Show if they were parallel
			if totalToolCalls > 1 {
				fmt.Printf("  🚀 Tools were executed in PARALLEL!\n")
			}
		}
	}

	// Check for tool calls in the final response
	if resp != nil && len(resp.Choices) > 0 {
		for i, choice := range resp.Choices {
			// Display any thinking content from GenerationInfo
			if choice.GenerationInfo != nil {
				if choice.Reasoning != nil && choice.Reasoning.Content != "" {
					fmt.Println("\n  📝 Captured Thinking Process:")
					fmt.Println("  ├─────────────────────────────")
					// Display first 500 chars of thinking
					thinkingPreview := choice.Reasoning.Content
					if len(thinkingPreview) > 500 {
						thinkingPreview = thinkingPreview[:500] + "..."
					}
					// Indent the thinking content
					lines := strings.Split(thinkingPreview, "\n")
					for _, line := range lines {
						fmt.Printf("  │ %s\n", line)
					}
					fmt.Println("  └─────────────────────────────")
				}
			}

			// Display tool calls
			if len(choice.ToolCalls) > 0 {
				fmt.Printf("\n  🔧 Tool Execution Summary:\n")
				fmt.Println("  ├─────────────────────────────")
				for j, tc := range choice.ToolCalls {
					fmt.Printf("  │ Call %d: %s\n", j+1, tc.FunctionCall.Name)
					fmt.Printf("  │   Args: %s\n", tc.FunctionCall.Arguments)
					// Simulate tool execution
					result := executeToolCall(tc.FunctionCall.Name, tc.FunctionCall.Arguments)
					fmt.Printf("  │   → %s\n", result)
					if j < len(choice.ToolCalls)-1 {
						fmt.Println("  │")
					}
				}
				fmt.Println("  └─────────────────────────────")
			}

			// Display content if any
			if choice.Content != "" {
				fmt.Printf("\n  💬 Final Response %d:\n", i+1)
				fmt.Println("  ├─────────────────────────────")
				// Indent the response content
				lines := strings.Split(choice.Content, "\n")
				for _, line := range lines {
					if line != "" {
						fmt.Printf("  │ %s\n", line)
					}
				}
				fmt.Println("  └─────────────────────────────")
			}
		}
	}

	// Stream statistics
	if streamedContent.Len() > 0 {
		fmt.Println("\n  📊 Stream Statistics:")
		fmt.Println("  ├─────────────────────────────")
		fmt.Printf("  │ Streamed Text:     %d bytes\n", streamedContent.Len())
		fmt.Printf("  │ Response Blocks:   %d\n", contentBlockCount)
		fmt.Println("  └─────────────────────────────")
	}

	// Stage 6: Token Metrics
	fmt.Println("\n📈 STAGE 6: TOKEN METRICS")
	fmt.Println("─────────────────────────")

	// Try to get token info from GenerationInfo or from captured streaming data
	var inputTokens, outputTokens, totalTokens int

	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		// Try different possible field names
		if v, ok := genInfo["PromptTokens"].(int); ok {
			inputTokens = v
		} else if v, ok := genInfo["InputTokens"].(int); ok {
			inputTokens = v
		} else if v, ok := genInfo["prompt_tokens"].(int); ok {
			inputTokens = v
		}

		if v, ok := genInfo["CompletionTokens"].(int); ok {
			outputTokens = v
		} else if v, ok := genInfo["OutputTokens"].(int); ok {
			outputTokens = v
		} else if v, ok := genInfo["completion_tokens"].(int); ok {
			outputTokens = v
		}

		if v, ok := genInfo["TotalTokens"].(int); ok {
			totalTokens = v
		} else if v, ok := genInfo["total_tokens"].(int); ok {
			totalTokens = v
		} else {
			totalTokens = inputTokens + outputTokens
		}
	}

	fmt.Println("  Token Usage Summary:")
	fmt.Println("  ├─────────────────────────────")
	fmt.Printf("  │ Input Tokens:      %d\n", inputTokens)
	fmt.Printf("  │ Output Tokens:     %d\n", outputTokens)
	fmt.Printf("  │ Total Tokens:      %d\n", totalTokens)
	fmt.Println("  └─────────────────────────────")

	// Extract thinking token details
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		if v, ok := genInfo["ReasoningTokens"].(int); ok && v > 0 {
			fmt.Println("\n  Interleaved Thinking Analysis:")
			fmt.Println("  ├─────────────────────────────")
			fmt.Printf("  │ Thinking Tokens:   %d\n", v)
			fmt.Printf("  │ Visible Output:    %d\n", outputTokens-v)
			fmt.Printf("  │ Thinking Ratio:    %.1f%% of output\n",
				float64(v)/float64(outputTokens)*100)
			fmt.Println("  └─────────────────────────────")

			fmt.Println("\n  💡 Thinking Benefits:")
			fmt.Println("  • Planning tool usage strategy")
			fmt.Println("  • Interpreting intermediate results")
			fmt.Println("  • Synthesizing multi-step analysis")
			fmt.Println("  • Ensuring logical coherence")
		}
	}

	// Stage 7: Cache Analytics
	fmt.Println("\n💰 STAGE 7: CACHE ANALYTICS")
	fmt.Println("─────────────────────────")

	if len(cacheStatsList) > 0 {
		fmt.Println("\n  Detailed Cache Statistics by Iteration:")
		fmt.Println("  ├─────────────────────────────────────────────────────────")

		totalCacheWrites := 0
		totalCacheReads := 0
		totalRegularInput := 0
		totalOutput := 0
		cumulativeEffectiveCost := 0.0
		cumulativeNoCacheCost := 0.0

		for i, stats := range cacheStatsList {
			fmt.Printf("  │ Iteration %d:\n", stats.Iteration)
			fmt.Printf("  │   Input Tokens:         %6d (regular: %d, cache write: %d, cache read: %d)\n",
				stats.InputTokens,
				stats.InputTokens,
				stats.CacheCreation,
				stats.CacheRead)
			fmt.Printf("  │   Output Tokens:        %6d\n", stats.OutputTokens)

			if stats.Cache1hCreation > 0 {
				fmt.Printf("  │   Cache Created (1h):   %6d tokens\n", stats.Cache1hCreation)
			}
			if stats.Cache5mCreation > 0 {
				fmt.Printf("  │   Cache Created (5m):   %6d tokens\n", stats.Cache5mCreation)
			}

			if i > 0 && stats.CacheRead > 0 {
				// Calculate savings for this iteration
				fmt.Printf("  │   Effective Cost:       %.0f tokens (vs %d without cache)\n",
					stats.EffectiveCost, stats.TotalTokens)
				fmt.Printf("  │   Savings:              %.1f%%\n", stats.SavingsVsNoCache)
			} else if stats.CacheCreation > 0 {
				fmt.Printf("  │   Effective Cost:       %.0f tokens (cache creation = 125%% cost)\n",
					stats.EffectiveCost)
				fmt.Printf("  │   Note:                 Initial cache write (investment for future savings)\n")
			}

			totalCacheWrites += stats.CacheCreation
			totalCacheReads += stats.CacheRead
			totalRegularInput += stats.InputTokens
			totalOutput += stats.OutputTokens
			cumulativeEffectiveCost += stats.EffectiveCost
			cumulativeNoCacheCost += float64(stats.TotalTokens)

			if i < len(cacheStatsList)-1 {
				fmt.Println("  │")
			}
		}

		fmt.Println("  └─────────────────────────────────────────────────────────")

		fmt.Println("\n  📊 Cumulative Cache Economics:")
		fmt.Println("  ├─────────────────────────────────────────────────────────")
		fmt.Printf("  │ Total Cache Writes:     %6d tokens @ 125%% = %.0f effective tokens\n",
			totalCacheWrites, float64(totalCacheWrites)*1.25)
		fmt.Printf("  │ Total Cache Reads:      %6d tokens @ 10%%  = %.0f effective tokens\n",
			totalCacheReads, float64(totalCacheReads)*0.10)
		fmt.Printf("  │ Regular Input:          %6d tokens @ 100%% = %d effective tokens\n",
			totalRegularInput, totalRegularInput)
		fmt.Printf("  │ Output Tokens:          %6d tokens @ 100%% = %d effective tokens\n",
			totalOutput, totalOutput)
		fmt.Println("  │")
		fmt.Printf("  │ Total Effective Cost:   %.0f tokens\n", cumulativeEffectiveCost)
		fmt.Printf("  │ Cost Without Cache:     %.0f tokens\n", cumulativeNoCacheCost)
		totalSavings := ((cumulativeNoCacheCost - cumulativeEffectiveCost) / cumulativeNoCacheCost) * 100
		fmt.Printf("  │ Total Savings:          %.1f%%\n", totalSavings)
		fmt.Println("  └─────────────────────────────────────────────────────────")

		fmt.Println("\n  💡 Cache Economics Explained:")
		fmt.Println("  ├─────────────────────────────────────────────────────────")
		fmt.Println("  │ Cache Write (1st time):  125% cost - One-time investment")
		fmt.Println("  │ Cache Read (subsequent): 10% cost - 90% savings!")
		fmt.Println("  │")
		fmt.Println("  │ With Interleaved Thinking + Tools:")
		fmt.Println("  │   • Tool definitions cached once, reused every iteration")
		fmt.Println("  │   • Thinking blocks from previous turns cached incrementally")
		fmt.Println("  │   • Each iteration reads old context (10%) + writes new content (125%)")
		fmt.Println("  │   • Result: Massive savings on multi-turn agent workflows!")
		fmt.Println("  └─────────────────────────────────────────────────────────")

		if len(cacheStatsList) > 1 {
			firstIterCost := cacheStatsList[0].EffectiveCost
			avgLaterIterCost := 0.0
			for i := 1; i < len(cacheStatsList); i++ {
				avgLaterIterCost += cacheStatsList[i].EffectiveCost
			}
			avgLaterIterCost /= float64(len(cacheStatsList) - 1)

			fmt.Println("\n  🎯 Key Insights:")
			fmt.Println("  ├─────────────────────────────────────────────────────────")
			fmt.Printf("  │ First Iteration:        %.0f effective tokens (cache creation)\n", firstIterCost)
			fmt.Printf("  │ Avg Subsequent Iters:   %.0f effective tokens (cache reuse)\n", avgLaterIterCost)
			iterSavings := ((firstIterCost - avgLaterIterCost) / firstIterCost) * 100
			fmt.Printf("  │ Per-Iteration Savings:  %.1f%% after first iteration\n", iterSavings)
			fmt.Println("  │")
			fmt.Println("  │ Conclusion: Cache pays for itself after ~2 iterations!")
			fmt.Println("  │             Longer conversations = exponential savings")
			fmt.Println("  └─────────────────────────────────────────────────────────")
		}
	} else {
		fmt.Println("  ⚠️  No cache statistics available")
	}

	// Final Summary
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    DEMO COMPLETE                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	fmt.Println("\n🎯 Key Features Demonstrated:")
	fmt.Println("  ✓ Interleaved thinking between tool calls")
	fmt.Println("  ✓ Real-time progress tracking with stages")
	fmt.Println("  ✓ Clear visualization of thinking blocks")
	fmt.Println("  ✓ Tool orchestration with results")
	fmt.Println("  ✓ Token usage analysis with thinking breakdown")
	fmt.Println("  ✓ Prompt caching with detailed economics")

	fmt.Println("\n📚 Use Cases:")
	fmt.Println("  • Complex multi-step analysis tasks")
	fmt.Println("  • Data processing with reasoning")
	fmt.Println("  • Research requiring tool coordination")
	fmt.Println("  • Decision-making with transparent logic")
	fmt.Println("  • Cost-effective multi-turn agent workflows")
}

// convertToolCallsToParts converts tool calls to content parts
func convertToolCallsToParts(toolCalls []llms.ToolCall) []llms.ContentPart {
	parts := make([]llms.ContentPart, 0, len(toolCalls))
	for _, tc := range toolCalls {
		parts = append(parts, llms.ToolCall{
			ID:           tc.ID,
			Type:         tc.Type,
			FunctionCall: tc.FunctionCall,
		})
	}
	return parts
}

// Simulate tool execution (in real use, these would call actual functions)
func executeToolCall(name string, arguments string) string {
	var args map[string]interface{}
	json.Unmarshal([]byte(arguments), &args)

	switch name {
	case "calculate":
		expr, _ := args["expression"].(string)
		// Simulate various calculations
		switch {
		case strings.Contains(expr, "19.2") && strings.Contains(expr, "16.1"):
			return "19.25"
		case strings.Contains(expr, "(17.4/15.3"):
			return "0.0534" // Average quarterly growth
		case strings.Contains(expr, "19.2") && strings.Contains(expr, "12.5"):
			return "53.6"
		case strings.Contains(expr, "12.5") && strings.Contains(expr, "14.2") && strings.Contains(expr, "/4"):
			return "14.15"
		case strings.Contains(expr, "15.3") && strings.Contains(expr, "17.4") && strings.Contains(expr, "/4"):
			return "17.2"
		default:
			// Generic calculation result
			return fmt.Sprintf("Result: %.2f", 15.5)
		}

	case "search_knowledge":
		query, _ := args["query"].(string)
		// Return different results based on query
		if strings.Contains(query, "seasonal") {
			return "Seasonal patterns show Q4 typically strongest due to holiday shopping, Q1 weakest post-holiday"
		} else if strings.Contains(query, "factors") {
			return "Key growth drivers: digital transformation, market expansion, improved supply chain efficiency"
		} else if strings.Contains(query, "economic") {
			return "Economic indicators: consumer confidence up 12%, GDP growth 2.8%, inflation moderating at 3.2%"
		}
		return fmt.Sprintf("Research data for '%s' retrieved", query)

	case "analyze_data":
		analysisType, _ := args["analysis_type"].(string)
		data, _ := args["data"].([]interface{})

		switch analysisType {
		case "trend":
			return "Trend: Consistent upward trajectory with 7.2% quarter-over-quarter average growth"
		case "std_dev":
			return "Standard Deviation: 2.41 million (moderate volatility)"
		case "mean":
			return "Mean: 15.675 million average across all quarters"
		default:
			return fmt.Sprintf("Analysis complete for %s with %d data points", analysisType, len(data))
		}

	case "make_prediction":
		baseValue, _ := args["base_value"].(float64)
		growthRate, _ := args["growth_rate"].(float64)
		seasonality, _ := args["seasonality_factor"].(float64)
		confidence, _ := args["confidence_level"].(string)

		// Calculate prediction
		prediction := baseValue * (1 + growthRate) * seasonality

		return fmt.Sprintf("Prediction: %.2f million (confidence: %s, growth: %.1f%%, seasonality: %.2fx)",
			prediction, confidence, growthRate*100, seasonality)

	case "generate_report":
		growthRate, _ := args["growth_rate"].(float64)
		trendDir, _ := args["trend_direction"].(string)
		volatility, _ := args["volatility_level"].(string)
		prediction, _ := args["prediction"].(float64)

		return fmt.Sprintf("Strategic Report Generated:\n"+
			"- YoY Growth: %.1f%%\n"+
			"- Trend: %s\n"+
			"- Volatility: %s\n"+
			"- Q1-2025 Forecast: %.2f million\n"+
			"- Recommendation: Continue growth strategy with focus on Q4 seasonality",
			growthRate, trendDir, volatility, prediction)

	default:
		return "Tool execution completed"
	}
}
