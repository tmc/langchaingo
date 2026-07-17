package bedrock

const (
	// Amazon Nova 2 Lite is an advanced multimodal model geared towards adaptive reasoning, efficient thinking,
	// customization and agentic workflows. It intelligently balances performance and efficiency by dynamically
	// adjusting reasoning depth based on task complexity.
	//
	// Max tokens: 1M
	// Languages: 200+ languages (optimized for English, German, Spanish, French, Italian, Japanese, Korean, Arabic, Simplified Chinese, Russian, Hindi, Portuguese, Dutch, Turkish, and Hebrew).
	ModelAmazonNova2LiteV1 = "us.amazon.nova-2-lite-v1:0"

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

	// Claude Fable 5 is Anthropic's most capable widely released model, built for the most demanding
	// reasoning and long-horizon agentic work. Thinking is always on (adaptive); the raw chain of thought
	// is never returned. On Bedrock it requires opting into the provider data-share retention mode, and
	// temperature must be 1.0 or unset.
	//
	// Max tokens: 1M
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeFable5 = "us.anthropic.claude-fable-5"

	// Claude Opus 4.8 is Anthropic's most capable Opus-tier model — highly autonomous, state-of-the-art
	// on long-horizon agentic work, knowledge work, and memory, with clearer and warmer writing.
	// Supports adaptive thinking only: budget thinking and sampling params are rejected.
	//
	// Max tokens: 1M
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeOpus48 = "us.anthropic.claude-opus-4-8"

	// Claude Opus 4.7 is a highly autonomous previous-generation Opus, strong on long-horizon agentic
	// work, knowledge work, vision, and memory. It introduces the xhigh effort level and high-resolution
	// vision. Supports adaptive thinking only: budget thinking and sampling params are rejected.
	//
	// Max tokens: 1M
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeOpus47 = "us.anthropic.claude-opus-4-7"

	// Claude Sonnet 5 is Anthropic's most capable Sonnet, built for coding, agents, and professional
	// work at scale with near-Opus intelligence at Sonnet cost. Adaptive thinking is on by default;
	// budget thinking and non-default sampling params are rejected. On Bedrock, adaptive thinking is
	// always on and cannot be disabled, unlike the Anthropic API where it accepts an explicit disable.
	//
	// Max tokens: 1M
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeSonnet5 = "us.anthropic.claude-sonnet-5"

	// Claude Opus 4.6 is the world's best model for coding, enterprise agents, and professional work.
	// It excels at agentic workflows, orchestrating complex tasks across dozens of tools with industry-leading
	// reliability. It handles the full lifecycle from architecture to deployment, delivers the deepest reasoning
	// for security workflows, and is Anthropic's most capable model for financial workflows and computer use.
	//
	// Max tokens: 1M
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeOpus46 = "us.anthropic.claude-opus-4-6-v1"

	// Claude Sonnet 4.6 delivers frontier intelligence at scale—built for coding, agents, and enterprise workflows.
	// It excels at complex, multi-step tasks requiring sustained reasoning and adaptive decision-making, handles
	// iterative development work with complex codebases, and brings professional-grade analysis with memory to
	// maintain context across files. Step-change improvement in creating spreadsheets, slides, and docs.
	//
	// Max tokens: 1M
	// Languages: English, French, Modern Standard Arabic, Mandarin Chinese, Hindi, Spanish, Portuguese, Korean, Japanese, German, Russian, Polish, and other languages.
	ModelAnthropicClaudeSonnet46 = "us.anthropic.claude-sonnet-4-6"

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

	// DeepSeek-V3.2 harmonizes high computational efficiency with superior reasoning and agent performance.
	// It builds on DeepSeek Sparse Attention for long-context efficiency, a scalable reinforcement learning framework,
	// and a large-scale agentic task synthesis pipeline. This model excels at long-context reasoning and agentic tasks,
	// efficiently handling extended inputs while maintaining strong accuracy. Its sparse attention design enables it to
	// process complex, multi-step workflows without excessive compute costs. Targets long-context reasoning,
	// tool-using agents, and efficient deployment in production environments.
	//
	// Max tokens: 164k
	// Languages: English, Chinese.
	ModelDeepSeekV32 = "deepseek.v3.2"

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

	// Qwen3 Next 80B A3B turns cutting-edge MoE and hybrid attention into a practical, ultra-long-context assistant
	// that scales from everyday chat to million-token workflows. It delivers flagship-level reasoning, coding,
	// and agent performance with only 3B active parameters per token. Ideal for long-context summarization,
	// code generation/refactoring, enterprise knowledge QA, and stable agentic workflows with tools.
	//
	// Max tokens: 256k
	// Languages: English, Chinese.
	ModelQwen3Next80BA3B = "qwen.qwen3-next-80b-a3b"

	// Qwen3 VL 235B A22B is a frontier vision-language model that sees, reads, and reasons across images, documents,
	// and long videos at massive scale. Its 235B-parameter MoE architecture (≈22B active) delivers state-of-the-art
	// multimodal understanding, OCR, and spatial reasoning over contexts reaching hundreds of thousands of tokens.
	// Ideal for document intelligence (OCR + layout), multimodal RAG, visual QA, and UI/scene understanding.
	//
	// Max tokens: 256k
	// Languages: English, Chinese.
	ModelQwen3VL235BA22B = "qwen.qwen3-vl-235b-a22b"

	// Qwen3 32B is a balanced dense model that offers strong reasoning and general-purpose performance with
	// straightforward deployment on standard infrastructure. Despite its smaller size compared to frontier-scale
	// models, Qwen3-32B delivers performance that surpasses many larger models, and proves highly versatile across
	// reasoning, coding, and research use cases. Its balance of capability, cost efficiency, and operational
	// simplicity has made it one of the most practical and widely deployable models in the Qwen3 family.
	//
	// Max tokens: 16384
	// Languages: English, Chinese.
	ModelQwen332BV1 = "qwen.qwen3-32b-v1:0"

	// Qwen3-Coder-30B-A3B-Instruct delivers strong coding and reasoning performance in a compact MoE design, making
	// it one of the most widely adopted models in the Qwen3-Coder series. It has become a favorite among developers
	// and enterprises seeking a practical balance between cost and capability. The model excels at "vibe coding,"
	// natural-language-first programming, debugging, SQL generation, and other development workflows, while being
	// lightweight enough to run on single high-memory GPUs or small clusters.
	//
	// Max tokens: 262144
	// Languages: English, Chinese.
	ModelQwen3Coder30BA3BV1 = "qwen.qwen3-coder-30b-a3b-v1:0"

	// Qwen3-Coder-Next is an open-weight language model built specifically for coding, with strong performance
	// on large-scale software engineering and agentic coding benchmarks. It uses a hybrid Mixture-of-Experts
	// architecture to offer high capability at relatively modest active parameter counts, improving efficiency
	// for real-world deployments. Optimized for tool use and function calling, making it suitable as the core
	// of coding agents that interact with shells, editors, issue trackers, and other developer tools.
	//
	// Max tokens: 256k
	// Languages: English, Chinese.
	ModelQwen3CoderNext = "qwen.qwen3-coder-next"

	// Mistral Large 3 is Mistral's most advanced open-weight multimodal model, combining a granular
	// Mixture-of-Experts architecture (673B total parameters with 39B active, plus a 2.5B vision encoder)
	// and a 256k context window to deliver state-of-the-art reliability, long-context reasoning, and
	// agentic performance for production assistants, RAG systems, scientific workloads, and complex enterprise applications.
	//
	// Max tokens: 256k
	// Languages: English, French, Spanish, German, Russian, Chinese, Japanese, Italian, Portuguese, Dutch, Polish, Vietnamese, Indonesian, Czech, Turkish, Farsi, Greek, Swedish, Arabic, Hungarian, Romanian, Finnish, Danish, Norwegian, Hebrew, Catalan, Hindi, Korean, Bengali, Tamil, Serbian, Urdu, Nepali, Marathi, Croatian, Telugu, Khmer, Tagalog, Gujarati, Malay, Kannada, Punjabi, Lao, Breton.
	ModelMistralLarge3 = "mistral.mistral-large-3-675b-instruct"

	// Devstral 2 123B is Mistral's 123-billion parameter (FP8) agentic model purpose-built for
	// software engineering: autonomous coding workflows, multi-file edits, and native tool-calling
	// to explore repositories and orchestrate complex engineering tasks. Scores 72.2% on SWE-bench
	// Verified and 61.3% on SWE-bench Multilingual.
	//
	// Max tokens: 256k
	// Languages: English (primary), French, Spanish, German, Italian, Portuguese, Chinese, Japanese, Korean, and 20+ additional languages.
	ModelMistralDevstral2123B = "mistral.devstral-2-123b"

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

	// Kimi K2.5 brings together strong vision, language, and code capabilities in a single natively multimodal
	// architecture. It handles complex tasks that mix images and text—such as generating code from UI mockups
	// or analyzing visual documents—with high accuracy. The model's "thinking" mode enables deep, deliberate reasoning,
	// while "instant" mode provides fast responses for interactive use. Its built-in support for tool use and agent
	// orchestration makes it highly effective for building sophisticated multimodal assistants.
	//
	// Max tokens: 256k
	// Languages: English, Chinese.
	ModelMoonshotKimiK25 = "moonshotai.kimi-k2.5"

	// Kimi K2 Thinking is Moonshot AI's flagship "thinking agent" model, designed for deep, tool-augmented reasoning.
	// Its 1T-parameter MoE architecture (32B active) powers state-of-the-art performance on long-horizon tasks like
	// HLE and BrowseComp. Native INT4 quantization and a 256K context window enable serious research- and agent-style
	// workloads with practical hardware. Ideal for long-horizon planning with tools, complex coding and debugging,
	// research agents over large corpora, and workflows needing 200-300-step stable tool orchestration.
	//
	// Max tokens: 256k
	// Languages: Multilingual (including Chinese and English).
	ModelMoonshotKimiK2Thinking = "moonshot.kimi-k2-thinking"

	// GLM-4.7 is a general-purpose language model in the GLM family with a focus on generating clean, modern
	// front-end code and web interfaces. It can turn natural-language descriptions into structured HTML, CSS,
	// and JavaScript while also supporting standard text and reasoning tasks. The model is positioned for developers
	// who want high-quality UI outputs alongside general conversational and coding abilities. It remains compatible
	// with typical LLM use cases such as question answering, summarization, and dialogue.
	//
	// Max tokens: 203k
	// Languages: English, Chinese.
	ModelGLM47 = "zai.glm-4.7"

	// GLM-4.7-Flash is a lightweight variant of GLM-4.7, using a mixture-of-experts architecture to reduce
	// resource requirements while maintaining strong output quality. It is designed for scenarios where low latency
	// and cost efficiency are important, such as interactive assistants or high-traffic services. The model retains
	// the core text and code generation capabilities of GLM-4.7 in a smaller active-parameter footprint.
	// A practical choice when deployment constraints limit the use of larger models.
	//
	// Max tokens: 203k
	// Languages: English, Chinese.
	ModelGLM47Flash = "zai.glm-4.7-flash"

	// GLM-5 is Z.ai's frontier reasoning and agentic model, a significant step up from the GLM-4.x line
	// on complex reasoning, coding, and multi-step agent workflows. It targets production agent systems
	// that need strong tool use and long-context planning.
	//
	// Max tokens: 200k
	// Languages: English, Chinese.
	ModelGLM5 = "zai.glm-5"

	// MiniMax M2.5 is an agent-native frontier model trained to reason efficiently, decompose tasks
	// optimally, and complete complex workflows under real-world time and cost constraints. Well
	// suited for production agents handling full-stack software projects, research and analysis
	// workflows, long-horizon planning, and multi-tool orchestration.
	//
	// Max tokens: 196k
	// Languages: English, Chinese.
	ModelMiniMaxM25 = "minimax.minimax-m2.5"

	// MiniMax M2.1 is an open-weight model focused on coding, tool use, and long-horizon task
	// planning, evaluated on practical front-end, backend, and workflow-automation benchmarks.
	// Intended as a general-purpose backbone for agent-based applications with improved reasoning,
	// coding, and instruction following over M2.
	//
	// Max tokens: 196k
	// Languages: English, Chinese.
	ModelMiniMaxM21 = "minimax.minimax-m2.1"

	// MiniMax M2 is a MoE model that blends frontier-level intelligence with highly efficient
	// active parameters, engineered for AI agents with strong reasoning, coding, and multilingual
	// performance at competitive cost. Suited for general-purpose chat/coding, tool-using agents,
	// multilingual assistants, and high-throughput inference.
	//
	// Max tokens: 400k
	// Languages: English, Chinese.
	ModelMiniMaxM2 = "minimax.minimax-m2"

	// NVIDIA Nemotron 3 Super 120B is an open hybrid mixture-of-experts model (about 12B active
	// parameters) built for reasoning, coding, and agentic tasks with strong cost efficiency.
	//
	// Max tokens: 256k
	// Languages: English.
	ModelNvidiaNemotronSuper3120B = "nvidia.nemotron-super-3-120b"
)
