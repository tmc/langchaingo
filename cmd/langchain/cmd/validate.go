package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate model connections and environment setup",
	Long: `Validate your LangChain Go environment by testing connections to various
LLM providers and checking configuration.

This command will:
  • Test API key validity
  • Verify model accessibility  
  • Check basic completion functionality
  • Validate environment setup

Supported providers:
  • OpenAI (GPT models)
  • Anthropic (Claude models)
  • More providers coming soon`,
	RunE: validateEnvironment,
}

var (
	validateProvider string
	quickTest        bool
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&validateProvider, "provider", "all", "Provider to validate (openai, anthropic, all)")
	validateCmd.Flags().BoolVar(&quickTest, "quick", false, "Run quick validation tests only")
}

func validateEnvironment(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 LangChain Go Environment Validation")
	fmt.Println(strings.Repeat("=", 40))

	var results []ValidationResult

	// Validate based on provider selection
	switch validateProvider {
	case "openai":
		results = append(results, validateOpenAI()...)
	case "anthropic":
		results = append(results, validateAnthropic()...)
	case "all":
		results = append(results, validateOpenAI()...)
		results = append(results, validateAnthropic()...)
	default:
		return fmt.Errorf("unsupported provider: %s", validateProvider)
	}

	// Print results
	fmt.Println("\n📊 Validation Results:")
	fmt.Println(strings.Repeat("-", 40))

	var passed, failed int
	for _, result := range results {
		status := "✅"
		if !result.Success {
			status = "❌"
			failed++
		} else {
			passed++
		}

		fmt.Printf("%s %s\n", status, result.Test)
		if result.Message != "" {
			fmt.Printf("   %s\n", result.Message)
		}
		if !result.Success && result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	fmt.Printf("\n📈 Summary: %d passed, %d failed\n", passed, failed)

	if failed > 0 {
		fmt.Println("\n💡 Troubleshooting tips:")
		fmt.Println("   • Check your API keys are set correctly")
		fmt.Println("   • Verify your API keys have proper permissions")
		fmt.Println("   • Check your network connection")
		fmt.Println("   • Review provider-specific requirements")
		return fmt.Errorf("validation failed with %d errors", failed)
	}

	fmt.Println("\n🎉 All validations passed! Your environment is ready.")
	return nil
}

type ValidationResult struct {
	Test    string
	Success bool
	Message string
	Error   string
}

func validateOpenAI() []ValidationResult {
	var results []ValidationResult

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		results = append(results, ValidationResult{
			Test:    "OpenAI API Key",
			Success: false,
			Error:   "OPENAI_API_KEY environment variable not set",
		})
		return results
	}

	results = append(results, ValidationResult{
		Test:    "OpenAI API Key",
		Success: true,
		Message: "Environment variable found",
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	llm, err := openai.New()
	if err != nil {
		results = append(results, ValidationResult{
			Test:    "OpenAI Client Initialization",
			Success: false,
			Error:   err.Error(),
		})
		return results
	}

	results = append(results, ValidationResult{
		Test:    "OpenAI Client Initialization",
		Success: true,
		Message: "Client created successfully",
	})

	// Test completion (unless quick mode)
	if !quickTest {
		testPrompt := "Say 'Hello from LangChain Go!'"
		response, err := llms.GenerateFromSinglePrompt(ctx, llm, testPrompt)

		if err != nil {
			results = append(results, ValidationResult{
				Test:    "OpenAI Completion Test",
				Success: false,
				Error:   err.Error(),
			})
		} else {
			results = append(results, ValidationResult{
				Test:    "OpenAI Completion Test",
				Success: true,
				Message: fmt.Sprintf("Response: %s", truncateString(response, 50)),
			})
		}
	}

	return results
}

func validateAnthropic() []ValidationResult {
	var results []ValidationResult

	// Check API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		results = append(results, ValidationResult{
			Test:    "Anthropic API Key",
			Success: false,
			Error:   "ANTHROPIC_API_KEY environment variable not set",
		})
		return results
	}

	results = append(results, ValidationResult{
		Test:    "Anthropic API Key",
		Success: true,
		Message: "Environment variable found",
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	llm, err := anthropic.New()
	if err != nil {
		results = append(results, ValidationResult{
			Test:    "Anthropic Client Initialization",
			Success: false,
			Error:   err.Error(),
		})
		return results
	}

	results = append(results, ValidationResult{
		Test:    "Anthropic Client Initialization",
		Success: true,
		Message: "Client created successfully",
	})

	// Test completion (unless quick mode)
	if !quickTest {
		testPrompt := "Say 'Hello from LangChain Go!'"
		response, err := llms.GenerateFromSinglePrompt(ctx, llm, testPrompt)

		if err != nil {
			results = append(results, ValidationResult{
				Test:    "Anthropic Completion Test",
				Success: false,
				Error:   err.Error(),
			})
		} else {
			results = append(results, ValidationResult{
				Test:    "Anthropic Completion Test",
				Success: true,
				Message: fmt.Sprintf("Response: %s", truncateString(response, 50)),
			})
		}
	}

	return results
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
