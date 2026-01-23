package bedrock

const (
	// Jamba 1.5 Large is part of the Jamba 1.5 Model Family with a 256K token effective context window,
	// one of the largest on the market. Jamba 1.5 models focus on speed and efficiency,
	// delivering up to 2.5x faster inference than leading models of comparable size.
	// Jamba supports function calling/tool use, structured output (JSON) and documents API.
	//
	// Popular use cases include text generation, conversation, and instruction following.
	//
	// Max tokens: 256k
	// Languages: English (primary), French, Spanish, Portuguese, German, Arabic, Hebrew, and many others.
	ModelAi21Jamba15LargeV1 = "ai21.jamba-1-5-large-v1:0"

	// Jamba 1.5 Mini is part of the Jamba 1.5 Model Family with a 256K token effective context window,
	// one of the largest on the market. Jamba 1.5 models focus on speed and efficiency,
	// delivering up to 2.5x faster inference than leading models of comparable size.
	// Jamba supports function calling/tool use, structured output (JSON) and documents API.
	//
	// Popular use cases include text generation, conversation, and instruction following.
	//
	// Max tokens: 256k
	// Languages: English (primary), French, Spanish, Portuguese, German, Arabic, Hebrew, and many others.
	ModelAi21Jamba15MiniV1 = "ai21.jamba-1-5-mini-v1:0"

	// Amazon Nova 2 Lite is an advanced multimodal model geared towards adaptive reasoning, efficient thinking,
	// customization and agentic workflows. It intelligently balances performance and efficiency by dynamically
	// adjusting reasoning depth based on task complexity.
	//
	// Max tokens: 1M
	// Languages: 200+ languages (optimized for English, German, Spanish, French, Italian, Japanese, Korean, Arabic, Simplified Chinese, Russian, Hindi, Portuguese, Dutch, Turkish, and Hebrew).
	ModelAmazonNova2LiteV1 = "us.amazon.nova-2-lite-v1:0"

	// Amazon Nova Premier is the most capable of Amazon's multimodal models for complex reasoning tasks
	// and for use as the best teacher for distilling custom models. It supports agents, chat optimization,
	// code generation, complex reasoning analysis, conversation, math, multilingual support,
	// question answering, RAG, text generation, text summarization, translation, and video-to-text.
	//
	// Max tokens: 1M
	// Languages: 200+ languages.
	ModelAmazonNovaPremiereV1 = "us.amazon.nova-premier-v1:0"

	// Amazon Nova Pro is a multimodal understanding foundation model. It is multilingual and can reason
	// over text, images and videos. It supports agents, chat optimization, code generation, complex
	// reasoning analysis, conversation, math, multilingual support, question answering, RAG, text
	// generation, text summarization, translation, and video-to-text.
	//
	// Max tokens: 300k
	// Languages: 200+ languages.
	ModelAmazonNovaProV1 = "us.amazon.nova-pro-v1:0"

	// Amazon Nova Lite is a multimodal understanding foundation model. It is multilingual and can reason
	// over text, images and videos. It supports agents, chat optimization, conversation, math, multilingual
	// support, question answering, RAG, text generation, text summarization, translation, and video-to-text.
	//
	// Max tokens: 300k
	// Languages: 200+ languages.
	ModelAmazonNovaLiteV1 = "us.amazon.nova-lite-v1:0"

	// Amazon Nova Micro is a text-to-text understanding foundation model. It is multilingual and can reason
	// over text. It supports agents, chat optimization, conversation, math, multilingual support, question
	// answering, RAG, text generation, text summarization, and translation.
	//
	// Max tokens: 128k
	// Languages: 200+ languages.
	ModelAmazonNovaMicroV1 = "us.amazon.nova-micro-v1:0"

	// Claude Opus 4.5 is the next generation of Anthropic's most intelligent model, an industry leader
	// across coding, agents, computer use, and enterprise workflows. It can confidently deliver multi-day
	// software development projects in hours, working independently with technical depth.
	//
	// Max tokens: 200k
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeOpus45 = "us.anthropic.claude-opus-4-5-20251101-v1:0"

	// Claude Haiku 4.5 delivers near-frontier performance for a wide range of use cases, and stands out
	// as one of the best coding and agent models—with the right speed and cost to power free products
	// and high-volume user experiences.
	//
	// Max tokens: 200k
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeHaiku45 = "us.anthropic.claude-haiku-4-5-20251001-v1:0"

	// Claude Sonnet 4.5 is Anthropic's most powerful model for powering real-world agents, with industry-leading
	// capabilities around coding and computer use. It is the ideal balance of performance and practicality
	// for most internal and external use cases.
	//
	// Max tokens: 200k
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeSonnet45 = "us.anthropic.claude-sonnet-4-5-20250929-v1:0"

	// Claude Opus 4.1 is the next generation of Anthropic's most powerful model yet, an industry leader
	// for coding. It delivers sustained performance on long-running tasks that require focused effort
	// and thousands of steps, significantly expanding what AI agents can solve.
	//
	// Max tokens: 200k
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeOpus41 = "us.anthropic.claude-opus-4-1-20250805-v1:0"

	// Claude Opus 4 is Anthropic's most intelligent model and is state-of-the-art for coding
	// and agent capabilities, especially agentic search. It excels for customers needing
	// frontier intelligence including advanced coding, AI agents, agentic search and research,
	// long-horizon tasks and complex problem solving, and content creation.
	//
	// Max tokens: 200k
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish and other languages.
	ModelAnthropicClaudeOpus4 = "us.anthropic.claude-opus-4-20250514-v1:0"

	// Claude Sonnet 4 balances impressive performance for coding with the right speed and cost
	// for high-volume use cases. It handles everyday development tasks, powers production-ready
	// AI assistants, performs efficient research, and generates large-scale content.
	//
	// Max tokens: 200k
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish and other languages.
	ModelAnthropicClaudeSonnet4 = "us.anthropic.claude-sonnet-4-20250514-v1:0"

	// Claude 3.7 Sonnet is Anthropic's most intelligent model to date and the first Claude model
	// to offer extended thinking—the ability to solve complex problems with careful, step-by-step
	// reasoning. It's state-of-the-art for coding and ideal for powering AI agents.
	//
	// Max tokens: 200k
	// Languages: English, Spanish, Japanese, and multiple other languages.
	ModelAnthropicClaude37Sonnet = "us.anthropic.claude-3-7-sonnet-20250219-v1:0"

	// Claude 3.5 Haiku is Anthropic's fastest and most cost-effective model, excelling at use cases
	// like code and test case generation, sub-agents, and user-facing chatbots.
	//
	// Max tokens: 200k
	// Languages: English, Spanish, Japanese, and multiple other languages.
	ModelAnthropicClaude35Haiku = "us.anthropic.claude-3-5-haiku-20241022-v1:0"

	// Command R is a generative language model optimized for long-context tasks and large scale production workloads.
	// It supports natural language processing, text generation, and text summarization.
	//
	// Max tokens: 128k
	// Languages: English, French, Spanish, Italian, German, Portuguese, Japanese, Korean, Arabic, and Chinese.
	ModelCohereCommandRV1 = "cohere.command-r-v1:0"

	// Command R+ is a highly performant generative language model optimized for large scale production workloads.
	// It supports natural language processing, text generation, and text summarization.
	//
	// Max tokens: 128k
	// Languages: English, French, Spanish, Italian, German, Portuguese, Japanese, Korean, Arabic, and Chinese.
	ModelCohereCommandRPlusV1 = "cohere.command-r-plus-v1:0"

	// Llama 4 Maverick offers unparalleled, industry-leading performance in image and text understanding
	// with support for 12 languages, enabling the creation of sophisticated AI applications that bridge
	// language barriers. As the product workhorse model for general assistant and chat use cases,
	// it's great for precise image understanding and creative writing.
	//
	// Max tokens: 1M
	// Languages: English, French, German, Hindi, Italian, Portuguese, Spanish, Thai, Arabic, Indonesian, Tagalog, Vietnamese.
	ModelMetaLlama4MaverickInstructV1 = "us.meta.llama4-maverick-17b-instruct-v1:0"

	// Llama 4 Scout is a general purpose model with 17 billion active parameters, 16 experts, and 109 billion
	// total parameters that delivers state-of-the-art performance for its class. Scout dramatically increases
	// the supported context length to an industry leading 10 million tokens, opening up possibilities for
	// multi-document summarization, parsing extensive user activity, and reasoning over vast codebases.
	//
	// Max tokens: 3.5M
	// Languages: English, French, German, Hindi, Italian, Portuguese, Spanish, Thai, Arabic, Indonesian, Tagalog, Vietnamese.
	ModelMetaLlama4ScoutInstructV1 = "us.meta.llama4-scout-17b-instruct-v1:0"

	// Llama 3.3 70B offers on par performance with the 405B model at a lower cost.
	// With tool use, code generation, advanced reasoning and decision making, and steerability.
	// We recommend upgrading to this model as soon as possible for optimal performance.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama3370bInstructV1 = "us.meta.llama3-3-70b-instruct-v1:0"

	// Llama 3.2 90B Vision Instruct is a multimodal, fine-tuned model with 90 billion parameters
	// that delivers unparalleled capabilities in image understanding, visual reasoning, and multimodal
	// interaction. It enables advanced applications such as image captioning, image-text retrieval,
	// visual grounding, visual question answering, and document visual question answering.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama3290bInstructV1 = "us.meta.llama3-2-90b-instruct-v1:0"

	// Llama 3.2 11B Vision Instruct is a multimodal, fine-tuned model with 11 billion parameters
	// that delivers unparalleled capabilities in image understanding, visual reasoning, and multimodal
	// interaction. It enables advanced applications such as image captioning, image-text retrieval,
	// visual grounding, visual question answering, and document visual question answering.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama3211bInstructV1 = "us.meta.llama3-2-11b-instruct-v1:0"

	// Llama 3.1 70B Instruct is an update to Meta Llama 3 70B Instruct that includes an expanded 128K context length,
	// multilinguality and improved reasoning capabilities. It's optimized for multilingual dialogue use cases
	// and outperforms many available open source chat models on common industry benchmarks.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama3170bInstructV1 = "us.meta.llama3-1-70b-instruct-v1:0"

	// Llama 3.1 8B Instruct is an update to Meta Llama 3 8B Instruct that includes an expanded 128K context length,
	// multilinguality and improved reasoning capabilities. It's optimized for multilingual dialogue use cases
	// and outperforms many available open source chat models on common industry benchmarks.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama318bInstructV1 = "meta.llama3-1-8b-instruct-v1:0"

	// Meta Llama 3 70B Instruct is an accessible, open large language model designed for developers,
	// researchers, and businesses to build, experiment, and responsibly scale their generative AI ideas.
	// Ideal for content creation, conversational AI, language understanding, R&D, and Enterprise applications.
	//
	// Max tokens: 8k
	// Languages: English.
	ModelMetaLlama370bInstructV1 = "meta.llama3-70b-instruct-v1:0"

	// Meta Llama 3 8B Instruct is an accessible, open large language model designed for developers,
	// researchers, and businesses to build, experiment, and responsibly scale their generative AI ideas.
	// Ideal for limited computational power and resources, edge devices, and faster training times.
	//
	// Max tokens: 8k
	// Languages: English.
	ModelMetaLlama38bInstructV1 = "meta.llama3-8b-instruct-v1:0"

	// DeepSeek-R1 provides customers a state-of-the-art reasoning model, optimized for general reasoning tasks,
	// math, science, and code generation. This model is created by DeepSeek and developed through a combination
	// of cold-start data and reinforcement learning. DeepSeek-R1 is a text-only model supporting English and Chinese.
	//
	// Max tokens: 128k
	// Languages: English, Chinese.
	ModelDeepSeekR1V1 = "us.deepseek.r1-v1:0"

	// OpenAI GPT-OSS-120B delivers performance comparable to and surpassing leading alternatives, particularly
	// in coding, scientific analysis, and mathematical reasoning tasks. It excels in intelligent automation,
	// software development, complex problem-solving, and scientific research applications.
	//
	// Max tokens: 128k
	// Languages: English.
	ModelOpenAIGptOss120BV1 = "openai.gpt-oss-120b-1:0"

	// OpenAI GPT-OSS-20B delivers performance comparable to and surpassing leading alternatives, particularly
	// in coding, scientific analysis, and mathematical reasoning tasks. It excels in intelligent automation,
	// software development, complex problem-solving, and scientific research applications.
	//
	// Max tokens: 128k
	// Languages: English.
	ModelOpenAIGptOss20BV1 = "openai.gpt-oss-20b-1:0"

	// Qwen3 Next 80B A3B is an 80B-parameter MixtureOfExperts (MOE) language model (3B active) optimized
	// for instruction following and ultra-long-context tasks. It delivers flagship-level reasoning, coding,
	// and agent performance with only 3B active parameters per token.
	//
	// Max tokens: 256k
	// Languages: English, Chinese.
	ModelQwen3Next80BA3B = "qwen.qwen3-next-80b-a3b"

	// Qwen3 VL 235B A22B is a 235B-parameter MoE vision-language flagship model (about 22B active) for
	// high-end multimodal tasks, combining strong text generation with advanced visual perception,
	// 2D/3D spatial grounding, long-context document/video understanding, and multilingual OCR over very long contexts.
	//
	// Max tokens: 256k
	// Languages: English, Chinese.
	ModelQwen3VL235BA22B = "qwen.qwen3-vl-235b-a22b"

	// Qwen3 32B (dense) is a balanced dense model offering strong reasoning and general-purpose performance
	// with simple, predictable deployment across standard infra. It delivers performance that surpasses many
	// larger models and proves highly versatile across reasoning, coding, and research use cases.
	//
	// Max tokens: 16384
	// Languages: English, Chinese.
	ModelQwen332BV1 = "qwen.qwen3-32b-v1:0"

	// Qwen3-Coder-30B-A3B-Instruct delivers strong coding and reasoning performance in a compact MoE design.
	// It excels at "vibe coding," natural-language-first programming, debugging, SQL generation, and other
	// development workflows, while being lightweight enough to run on single high-memory GPUs or small clusters.
	//
	// Max tokens: 262144
	// Languages: English, Chinese.
	ModelQwen3Coder30BA3BV1 = "qwen.qwen3-coder-30b-a3b-v1:0"

	// Mistral Large 3 is Mistral's most advanced open-weight multimodal model, combining a granular
	// Mixture-of-Experts architecture (673B total parameters with 39B active, plus a 2.5B vision encoder)
	// and a 256k context window to deliver state-of-the-art reliability, long-context reasoning, and
	// agentic performance for production assistants, RAG systems, scientific workloads, and complex enterprise applications.
	//
	// Max tokens: 256k
	// Languages: English, French, Spanish, German, Russian, Chinese, Japanese, Italian, Portuguese, Dutch, Polish, Vietnamese, Indonesian, Czech, Turkish, Farsi, Greek, Swedish, Arabic, Hungarian, Romanian, Finnish, Danish, Norwegian, Hebrew, Catalan, Hindi, Korean, Bengali, Tamil, Serbian, Urdu, Nepali, Marathi, Croatian, Telugu, Khmer, Tagalog, Gujarati, Malay, Kannada, Punjabi, Lao, Breton.
	ModelMistralLarge3 = "mistral.mistral-large-3-675b-instruct"

	// Magistral Small 2509 is Mistral's small-sized dense model optimized for fast, cost-efficient instruction
	// following, reasoning, and coding, designed as a production-friendly "small but capable" assistant.
	// It brings "big model" quality to a smaller form factor with multimodal support for vision and text.
	//
	// Max tokens: 128k
	// Languages: English, French, German, Greek, Hindi, Indonesian, Italian, Japanese, Korean, Malay, Nepali, Polish, Portuguese, Romanian, Russian, Serbian, Spanish, Turkish, Ukrainian, Vietnamese, Arabic, Bengali, Chinese, and Farsi (24 languages total).
	ModelMistralMagistralSmall2509 = "mistral.magistral-small-2509"

	// Mistral Large (24.02) is the most advanced Mistral AI Large Language model capable of handling any language task
	// including complex multilingual reasoning, text understanding, transformation, and code generation.
	//
	// Max tokens: 32k
	// Languages: English, French, German, Spanish, Chinese, Japanese, and multiple other languages.
	ModelMistralLarge2402V1 = "mistral.mistral-large-2402-v1:0"

	// Kimi K2 Thinking is Moonshot AI's flagship "thinking agent" model, designed for deep, tool-augmented reasoning.
	// Its 1T-parameter MoE architecture (32B active) powers state-of-the-art performance on long-horizon tasks.
	// Native INT4 quantization and a 256K context window enable serious research- and agent-style workloads with practical hardware.
	//
	// Max tokens: 256k
	// Languages: Multilingual (including Chinese and English).
	ModelMoonshotKimiK2Thinking = "moonshot.kimi-k2-thinking"
)
