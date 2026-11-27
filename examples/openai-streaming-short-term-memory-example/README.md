# Adding Short-Term Memory to LLM Chatbots with LangChain Go

## Overview 📖

This contribution adds short-term memory functionality to LLM-powered chatbots using the LangChain Go library. It allows the chatbot to retain context from previous interactions, enhancing its ability to provide coherent and relevant responses. This implementation is designed to manage conversation context and message history effectively.

## Features 🚀

* **Short-Term Memory Management:** Implements a `Memory` struct to store and manage conversation history.
* **Context Handling:** Utilizes `llms.MessageContent` to maintain message roles (system, human, AI) and content.
* **Context Limiting:** Restricts the memory to a predefined `shortTermMemory` size, preventing excessive context buildup.
* **Integration with LangChain Go:** Seamlessly integrates with the LangChain Go library and OpenAI LLMs.
* **Streaming Support:** Uses `llms.WithStreamingFunc` to display responses in real-time.
* **Parameter Control:** Shows examples of using `llms.WithMaxTokens` and `llms.WithTemperature` to control LLM behavior.

## How It Works ⚙️

1.  **Memory Struct:** A `Memory` struct is created to store `llms.MessageContent` objects, representing the conversation history.
2.  **Adding Messages:** The `AddMessage` method appends new messages to the `Memory` struct, respecting the `shortTermMemory` limit.
3.  **Context Retrieval:** The `GetContext` method retrieves the current conversation context as a slice of `llms.MessageContent`.
4.  **LLM Interaction:** The application interacts with the OpenAI LLM using `llm.GenerateContent`, providing the conversation context.
5.  **Streaming Output:** The `showResponse` function prints the LLM's streaming output to the console.
6.  **Context Update:** LLM responses are added to the memory to keep track of the conversation.

## How to Use 🦫

``` golang

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Bem-vindo ao Assistant!")

	for {
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if strings.ToLower(input) == "exit" {
					fmt.Println("bye (waving hand)...")
					break
			}

			yourChatBotFunc(input)
			fmt.Println()
    }
}
```


## How to Run 🏃

1. Ensure you have Go installed.
2. Install the necessary dependencies using go get.
3. Set your OpenAI API key as an environment variable (OPENAI_API_KEY).
4. Run the Go program.