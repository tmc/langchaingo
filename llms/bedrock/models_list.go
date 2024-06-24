package bedrock

const (
	// Jurassic-2 Ultra is AI21’s most powerful model for complex tasks that require
	// advanced text generation and comprehension.
	//
	// Popular use cases include question answering, summarization, long-form copy generation,
	// advanced information extraction, and more.
	//
	// Max tokens: 8191
	//
	// Languages: English, Spanish, French, German, Portuguese, Italian, Dutch.
	ModelAi21J2UltraV1 = "ai21.j2-ultra-v1"

	// Jurassic-2 Mid is less powerful than Ultra, yet carefully designed to strike
	// the right balance between exceptional quality and affordability.
	//
	// Jurassic-2 Mid can be applied to any language comprehension or generation task
	// including question answering, summarization, long-form copy generation,
	// advanced information extraction and many others.
	//
	// Max tokens: 8191
	// Languages: English, Spanish, French, German, Portuguese, Italian, Dutch.
	ModelAi21J2MidV1 = "ai21.j2-mid-v1"
	// Amazon Titan Text Lite is a light weight efficient model ideal for fine-tuning
	// for English-language tasks, including like summarization and copywriting,
	// where customers want a smaller, more cost-effective model that is also highly customizable.
	//
	// Max tokens: 4k
	// Languages: English.
	ModelAmazonTitanTextLiteV1 = "amazon.titan-text-lite-v1"

	// Amazon Titan Text Express has a context length of up to 8,000 tokens,
	// making it well-suited for a wide range of advanced, general language tasks such as
	// open-ended text generation and conversational chat, as well as support within
	// Retrieval Augmented Generation (RAG).
	//
	// Max tokens: 8k
	// Languages: English (GA), Multilingual in 100+ languages (Preview).
	ModelAmazonTitanTextExpressV1 = "amazon.titan-text-express-v1"

	// Claude 3 Sonnet by Anthropic strikes the ideal balance between intelligence and
	// speed—particularly for enterprise workloads. It offers maximum utility at a lower
	// price than competitors, and is engineered to be the dependable, high-endurance
	// workhorse for scaled AI deployments.
	//
	// Claude 3 Sonnet can process images and return text outputs, and features a 200K context window.
	//
	// Max tokens: 200k
	// Languages: English and multiple other languages.
	ModelAnthropicClaudeV3Sonnet = "anthropic.claude-3-sonnet-20240229-v1:0"

	// Claude 3 Haiku is Anthropic's fastest, most compact model for near-instant responsiveness.
	// It answers simple queries and requests with speed.
	// Customers will be able to build seamless AI experiences that mimic human interactions.
	// Claude 3 Haiku can process images and return text outputs, and features a 200K context window.
	//
	// Max tokens: 200k
	// Languages: English and multiple other languages.
	ModelAnthropicClaudeV3Haiku = "anthropic.claude-3-haiku-20240307-v1:0"

	// ModelAnthropicClaudeV35Haiku
	// See more model info: https://docs.anthropic.com/en/docs/about-claude/models
	ModelAnthropicClaudeV35Haiku = "anthropic.claude-3-5-sonnet-20240620-v1:0"

	// An update to Claude 2 that features double the context window, plus improvements
	// across reliability, hallucination rates, and evidence-based accuracy in long document and RAG contexts.
	//
	// Max tokens: 200k
	// Languages: English and multiple other languages.
	ModelAnthropicClaudeV21 = "anthropic.claude-v2:1"

	// Anthropic's highly capable model across a wide range of tasks from sophisticated dialogue
	// and creative content generation to detailed instruction following.
	//
	// Max tokens: 100k
	// Languages: English and multiple other languages.
	ModelAnthropicClaudeV2 = "anthropic.claude-v2"

	// A fast, affordable yet still very capable model, which can handle a range of tasks
	// including casual dialogue, text analysis, summarization, and document question-answering.
	//
	// Max tokens: 100k
	// Languages: English and multiple other languages.
	ModelAnthropicClaudeInstantV1 = "anthropic.claude-instant-v1"

	// Command is Cohere's flagship text generation model.
	// It is trained to follow user commands and to be instantly useful in practical business applications.
	//
	// Max tokens: 4000
	// Languages: English.
	ModelCohereCommandTextV14 = "cohere.command-text-v14"

	// Cohere's Command-Light is a generative model that responds well with instruction-like prompts.
	// This model provides customers with an unbeatable balance of quality, cost-effectiveness, and low-latency inference.
	//
	// Max tokens: 4000
	// Languages: English.
	ModelCohereCommandLightTextV14 = "cohere.command-light-text-v14"

	// A dialogue use case optimized variant of Llama 2 models.
	// Llama 2 is an auto-regressive language model that uses an optimized transformer architecture.
	// Llama 2 is intended for commercial and research use in English.
	// This is the 13 billion parameter variant.
	//
	// Max tokens: 4096
	// Languages: English.
	ModelMetaLlama213bChatV1 = "meta.llama2-13b-chat-v1"

	// A dialogue use case optimized variant of Llama 2 models.
	// Llama 2 is an auto-regressive language model that uses an optimized transformer architecture.
	// Llama 2 is intended for commercial and research use in English.
	// This is the 70 billion parameter variant.
	//
	// Max tokens: 4096
	// Languages: English.
	ModelMetaLlama270bChatV1 = "meta.llama2-70b-chat-v1"

	// Llama 3 is the most capable Llama model yet, which supports a 8K context length that doubles the capacity of Llama 2.
	// Llama 3 was pretrained on over 15 trillion tokens of data from publicly available source.
	// This is the 8 billion parameter variant.
	//
	// Max tokens: 8k
	// Languages: English(Over 5% of the Llama 3 pretraining dataset consists of high-quality non-English data that covers over 30 languages.
	// However, we do not expect the same level of performance in these languages as in English.)
	ModelMetaLlama38bInstructV1 = "meta.llama3-8b-instruct-v1:0"

	// Llama 3 is the most capable Llama model yet, which supports a 8K context length that doubles the capacity of Llama 2.
	// Llama 3 was pretrained on over 15 trillion tokens of data from publicly available source.
	// This is the 70 billion parameter variant.
	//
	// Max tokens: 8k
	// Languages: English(Over 5% of the Llama 3 pretraining dataset consists of high-quality non-English data that covers over 30 languages.
	// However, we do not expect the same level of performance in these languages as in English.)
	ModelMetaLlama370bInstructV1 = "meta.llama3-70b-instruct-v1:0"
)
