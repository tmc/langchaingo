---
title: LLMs
---

# LLMs

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/model_io/llms)
</Info>

LangChainGo provides a shared interface for interacting with large language model providers.

All LLMs implement the `llms.Model` interface:

```go
// Model is an interface multi-modal models implement.
type Model interface {
	// GenerateContent asks the model to generate content from a sequence of
	// messages. It's the most general interface for multi-modal LLMs that support
	// chat-like interactions.
	GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)

	// Call is a simplified interface for a text-only Model, generating a single
	// string response from a single string prompt.
	//
	// Deprecated: this method is retained for backwards compatibility. Use the
	// more general [GenerateContent] instead.
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)
}
```

The interface provides two methods:
- **`GenerateContent`**: The modern, recommended method that supports multi-modal inputs and complex message sequences
- **`Call`**: A legacy method for simple text-to-text generation (deprecated but still supported)

For backwards compatibility, `llms.LLM` is provided as a type alias:

```go
// LLM is an alias for model, for backwards compatibility.
// Deprecated: This alias may be removed in the future; please use Model instead.
type LLM = Model
```

- [Fake LLM](Integrations/fake.md)
- [Groq](Integrations/groq.md)
- [Hugging Face](Integrations/huggingface.md)
- [Llamafile](Integrations/llamafile.md)
- [Local](Integrations/local.md)
- [Mistral](Integrations/mistral.md)
- [OpenAI](Integrations/openai.md)
- [Vertex AI](Integrations/vertexai.md)
- [watsonx](Integrations/watsonx.md)
