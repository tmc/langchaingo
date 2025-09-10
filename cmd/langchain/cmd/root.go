package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "langchain",
	Short: "LangChain Go CLI - Build AI applications with composable components",
	Long: `🦜️🔗 LangChain Go CLI

Build powerful AI applications with composable components using the official
LangChain Go CLI. This tool helps you explore, build, and validate AI applications
with ease.

🚀 COMMANDS:
  examples    Browse and run 60+ examples (LLMs, vector stores, agents, etc.)
  init        Scaffold new projects from templates
  validate    Test model connections and environment setup

💡 EXAMPLES:
  langchain examples list                    # List all available examples
  langchain examples run openai-completion   # Run a specific example
  langchain init my-bot --template chat-bot  # Create new chat bot project
  langchain validate --provider openai       # Test OpenAI connection

📚 DOCUMENTATION:
  Website: https://tmc.github.io/langchaingo/docs/
  GitHub:  https://github.com/tmc/langchaingo
  Discord: https://discord.gg/t9UbBQs2rG

Get started by exploring examples or creating your first project!`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
