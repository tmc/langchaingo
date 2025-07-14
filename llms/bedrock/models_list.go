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

	// Jamba-Instruct offers an impressive 256K context window and delivers the best value per price
	// on core text generation, summarization, and question answering tasks for the enterprise.
	//
	// Popular use cases include text generation, document summarization, and question answering.
	//
	// Max tokens: 256k
	// Languages: English.
	ModelAi21JambaInstructV1 = "ai21.jamba-instruct-v1:0"

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

	// Amazon Titan Text Premier is a powerful and advanced model within the Titan Text family,
	// designed to deliver superior performance across a wide range of enterprise applications.
	// With its cutting-edge capabilities, it offers enhanced accuracy and exceptional results,
	// making it an excellent choice for organizations seeking top-notch text processing solutions.
	//
	// Max tokens: 32k
	// Languages: English.
	ModelAmazonTitanTextPremierV1 = "amazon.titan-text-premier-v1:0"

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

	// Claude 3.5 Sonnet (v2) is the upgraded version that is now state-of-the-art for a variety of tasks
	// including real-world software engineering, agentic capabilities and computer use.
	// It can process images and return text outputs, and features a 200K context window.
	//
	// Max tokens: 200k
	// Languages: English, Spanish, Japanese, and multiple other languages.
	ModelAnthropicClaude35SonnetV2 = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"

	// Claude 3.5 Sonnet (v1) raises the industry bar for intelligence, outperforming competitor models
	// and Claude 3 Opus on a wide range of evaluations, with the speed and cost of the mid-tier model.
	// It can process images and return text outputs, and features a 200K context window.
	//
	// Max tokens: 200k
	// Languages: English, Spanish, Japanese, and multiple other languages.
	ModelAnthropicClaude35SonnetV1 = "us.anthropic.claude-3-5-sonnet-20240620-v1:0"

	// Claude 3 Opus is Anthropic's most powerful AI model, with state-of-the-art performance on highly complex tasks.
	// It can navigate open-ended prompts and sight-unseen scenarios with remarkable fluency and human-like understanding.
	// Claude 3 Opus can process images and return text outputs, and features a 200K context window.
	//
	// Max tokens: 200k
	// Languages: English and multiple other languages.
	ModelAnthropicClaude3Opus = "us.anthropic.claude-3-opus-20240229-v1:0"

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

	// Llama 3.2 11B Vision Instruct is a multimodal, fine-tuned model with 11 billion parameters
	// that delivers unparalleled capabilities in image understanding, visual reasoning, and multimodal
	// interaction. It enables advanced applications such as image captioning, image-text retrieval,
	// visual grounding, visual question answering, and document visual question answering.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama3211bInstructV1 = "us.meta.llama3-2-11b-instruct-v1:0"

	// Llama 3.1 8B Instruct is an update to Meta Llama 3 8B Instruct that includes an expanded 128K context length,
	// multilinguality and improved reasoning capabilities. It's optimized for multilingual dialogue use cases
	// and outperforms many available open source chat models on common industry benchmarks.
	//
	// Max tokens: 128k
	// Languages: English, German, French, Italian, Portuguese, Hindi, Spanish, and Thai.
	ModelMetaLlama318bInstructV1 = "us.meta.llama3-1-8b-instruct-v1:0"

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
)
