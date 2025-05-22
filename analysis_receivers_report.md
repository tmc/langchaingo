# Receiver Analysis Report

## Summary
- Total types analyzed: 148
- Total methods analyzed: 808

- Types using only value receivers: 84
- Types using only pointer receivers: 53
- Types with mixed receivers: 11

## Detailed Analysis

### AIChatMessage
- **Struct size**: 4 fields
- **File**: ../llms/chat_messages.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 3
  - `GetType()` - value receiver
  - `GetContent()` - value receiver
  - `GetFunctionCall()` - value receiver

### AIMessagePromptTemplate
- **Struct size**: 1 fields
- **File**: ../prompts/message_prompt_template.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `FormatMessages()` - value receiver
  - `GetInputVariables()` - value receiver

### API
- **Struct size**: 1 fields
- **File**: ../tools/metaphor/metaphor.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 8
  - `Name()` - pointer receiver
  - `Description()` - pointer receiver
  - `Call()` - pointer receiver
  - `performSearch()` - pointer receiver
  - `findSimilar()` - pointer receiver
  - `getContents()` - pointer receiver
  - `formatResults()` - pointer receiver
  - `formatContents()` - pointer receiver

### APIChain
- **Struct size**: 3 fields
- **File**: ../chains/api.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `runRequest()` - value receiver

### AgentFinalStreamHandler
- **Struct size**: 5 fields
- **File**: ../callbacks/agent_final_stream.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 3
  - `GetEgress()` - pointer receiver
  - `ReadFromEgress()` - pointer receiver
  - `HandleStreamingFunc()` - pointer receiver

### AllFilter
- **Struct size**: 1 fields
- **File**: ../vectorstores/bedrockknowledgebases/options.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `isFilter()` - value receiver

### AnyFilter
- **Struct size**: 1 fields
- **File**: ../vectorstores/bedrockknowledgebases/options.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `isFilter()` - value receiver

### AssemblyAIAudioTranscriptLoader
- **Struct size**: 5 fields
- **File**: ../documentloaders/assemblyai.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 4
  - `Load()` - pointer receiver
  - `transcribe()` - pointer receiver
  - `formatTranscript()` - pointer receiver
  - `LoadAndSplit()` - pointer receiver

### BaseIndex
- **Struct size**: 5 fields
- **File**: ../vectorstores/cloudsql/vectorstore.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `indexOptions()` - pointer receiver
  - `indexOptions()` - pointer receiver

### Bedrock
- **Struct size**: 4 fields
- **File**: ../embeddings/bedrock/bedrock.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `EmbedDocuments()` - pointer receiver
  - `EmbedQuery()` - pointer receiver

### BinaryContent
- **Struct size**: 2 fields
- **File**: ../llms/generatecontent.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 4
  - `String()` - value receiver
  - `isPart()` - value receiver
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### BooleanParser
- **Struct size**: 2 fields
- **File**: ../outputparser/boolean_parser.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `GetFormatInstructions()` - value receiver
  - `parse()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `Type()` - value receiver

### CSV
- **Struct size**: 2 fields
- **File**: ../documentloaders/csv.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `Load()` - value receiver
  - `LoadAndSplit()` - value receiver

### Cacher
- **Struct size**: 2 fields
- **File**: ../llms/cache/cache.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver

### Calculator
- **Struct size**: 1 fields
- **File**: ../tools/calculator.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `Description()` - value receiver
  - `Name()` - value receiver
  - `Call()` - value receiver

### ChatMessage
- **Struct size**: 8 fields
- **File**: ../llms/openai/internal/openaiclient/chat.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 4
  - `GetType()` - value receiver
  - `GetContent()` - value receiver
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### ChatMessageHistory
- **Struct size**: 5 fields
- **File**: ../memory/zep/zep_chat_history.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 38
  - `validateTable()` - pointer receiver
  - `addMessage()` - pointer receiver
  - `AddMessage()` - pointer receiver
  - `AddAIMessage()` - pointer receiver
  - `AddUserMessage()` - pointer receiver
  - `Clear()` - pointer receiver
  - `AddMessages()` - pointer receiver
  - `Messages()` - pointer receiver
  - `SetMessages()` - pointer receiver
  - `Messages()` - pointer receiver
  - `AddAIMessage()` - pointer receiver
  - `AddUserMessage()` - pointer receiver
  - `Clear()` - pointer receiver
  - `AddMessage()` - pointer receiver
  - `SetMessages()` - pointer receiver
  - `validateTable()` - pointer receiver
  - `addMessage()` - pointer receiver
  - `AddMessage()` - pointer receiver
  - `AddAIMessage()` - pointer receiver
  - `AddUserMessage()` - pointer receiver
  - `Clear()` - pointer receiver
  - `AddMessages()` - pointer receiver
  - `Messages()` - pointer receiver
  - `SetMessages()` - pointer receiver
  - `Messages()` - pointer receiver
  - `AddAIMessage()` - pointer receiver
  - `AddUserMessage()` - pointer receiver
  - `Clear()` - pointer receiver
  - `AddMessage()` - pointer receiver
  - `SetMessages()` - pointer receiver
  - `messagesFromZepMessages()` - pointer receiver
  - `messagesToZepMessages()` - pointer receiver
  - `Messages()` - pointer receiver
  - `AddAIMessage()` - pointer receiver
  - `AddUserMessage()` - pointer receiver
  - `Clear()` - pointer receiver
  - `AddMessage()` - pointer receiver
  - `SetMessages()` - pointer receiver

### ChatMessageModel
- **Struct size**: 2 fields
- **File**: ../llms/chat_messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `ToChatMessage()` - value receiver

### ChatPromptTemplate
- **Struct size**: 2 fields
- **File**: ../prompts/chat_prompt_template.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 4
  - `FormatPrompt()` - value receiver
  - `Format()` - value receiver
  - `FormatMessages()` - value receiver
  - `GetInputVariables()` - value receiver

### ChatPromptValue
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `String()` - value receiver
  - `Messages()` - value receiver

### Client
- **Struct size**: 1 fields
- **File**: ../tools/zapier/internal/client.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 59
  - `CreateCompletion()` - pointer receiver
  - `CreateMessage()` - pointer receiver
  - `setHeaders()` - pointer receiver
  - `do()` - pointer receiver
  - `decodeError()` - pointer receiver
  - `setCompletionDefaults()` - pointer receiver
  - `createCompletion()` - pointer receiver
  - `setMessageDefaults()` - pointer receiver
  - `createMessage()` - pointer receiver
  - `CreateCompletion()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Summarize()` - pointer receiver
  - `CreateGeneration()` - pointer receiver
  - `createChat()` - pointer receiver
  - `CreateCompletion()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `getAccessToken()` - pointer receiver
  - `CreateChat()` - pointer receiver
  - `buildURL()` - pointer receiver
  - `setHeaders()` - pointer receiver
  - `createEmbedding()` - pointer receiver
  - `RunInference()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `runInference()` - pointer receiver
  - `do()` - pointer receiver
  - `stream()` - pointer receiver
  - `sendHTTPRequest()` - pointer receiver
  - `processResponse()` - pointer receiver
  - `Generate()` - pointer receiver
  - `GenerateChat()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `createCompletion()` - pointer receiver
  - `CreateCompletion()` - pointer receiver
  - `stream()` - pointer receiver
  - `Generate()` - pointer receiver
  - `do()` - pointer receiver
  - `stream()` - pointer receiver
  - `Generate()` - pointer receiver
  - `GenerateChat()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `createChat()` - pointer receiver
  - `setCompletionDefaults()` - pointer receiver
  - `createCompletion()` - pointer receiver
  - `createEmbedding()` - pointer receiver
  - `CreateCompletion()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `CreateChat()` - pointer receiver
  - `setHeaders()` - pointer receiver
  - `buildURL()` - pointer receiver
  - `buildAzureURL()` - pointer receiver
  - `newRequest()` - pointer receiver
  - `Search()` - pointer receiver
  - `SetMaxResults()` - pointer receiver
  - `formatResults()` - pointer receiver
  - `Search()` - pointer receiver
  - `List()` - pointer receiver
  - `Execute()` - pointer receiver
  - `ExecuteAsString()` - pointer receiver

### ClientOptions
- **Struct size**: 4 fields
- **File**: ../tools/zapier/internal/client.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `Validate()` - pointer receiver

### Combining
- **Struct size**: 1 fields
- **File**: ../outputparser/combining.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `GetFormatInstructions()` - value receiver
  - `parse()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `Type()` - value receiver

### CombiningHandler
- **Struct size**: 1 fields
- **File**: ../callbacks/combining.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 16
  - `HandleText()` - value receiver
  - `HandleLLMStart()` - value receiver
  - `HandleLLMGenerateContentStart()` - value receiver
  - `HandleLLMGenerateContentEnd()` - value receiver
  - `HandleChainStart()` - value receiver
  - `HandleChainEnd()` - value receiver
  - `HandleToolStart()` - value receiver
  - `HandleToolEnd()` - value receiver
  - `HandleAgentAction()` - value receiver
  - `HandleAgentFinish()` - value receiver
  - `HandleRetrieverStart()` - value receiver
  - `HandleRetrieverEnd()` - value receiver
  - `HandleStreamingFunc()` - value receiver
  - `HandleChainError()` - value receiver
  - `HandleLLMError()` - value receiver
  - `HandleToolError()` - value receiver

### CommaSeparatedList
- **Struct size**: 0 fields
- **File**: ../outputparser/comma_seperated_list.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 4
  - `GetFormatInstructions()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `Type()` - value receiver

### ConditionalPromptSelector
- **Struct size**: 2 fields
- **File**: ../chains/prompt_selector.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `GetPrompt()` - value receiver

### Constitutional
- **Struct size**: 7 fields
- **File**: ../chains/constitutional.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 10
  - `Call()` - pointer receiver
  - `processCritiquesAndRevisions()` - pointer receiver
  - `GetMemory()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver
  - `Call()` - pointer receiver
  - `processCritiquesAndRevisions()` - pointer receiver
  - `GetMemory()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver

### ContainsFilter
- **Struct size**: 2 fields
- **File**: ../vectorstores/bedrockknowledgebases/options.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `isFilter()` - value receiver

### ConversationBuffer
- **Struct size**: 7 fields
- **File**: ../memory/buffer.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `MemoryVariables()` - pointer receiver
  - `LoadMemoryVariables()` - pointer receiver
  - `SaveContext()` - pointer receiver
  - `Clear()` - pointer receiver
  - `GetMemoryKey()` - pointer receiver

### ConversationTokenBuffer
- **Struct size**: 2 fields
- **File**: ../memory/token_buffer.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `MemoryVariables()` - pointer receiver
  - `LoadMemoryVariables()` - pointer receiver
  - `SaveContext()` - pointer receiver
  - `Clear()` - pointer receiver
  - `getNumTokensFromMessages()` - pointer receiver

### ConversationWindowBuffer
- **Struct size**: 1 fields
- **File**: ../memory/window_buffer.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `MemoryVariables()` - pointer receiver
  - `LoadMemoryVariables()` - pointer receiver
  - `SaveContext()` - pointer receiver
  - `cutMessages()` - pointer receiver
  - `Clear()` - pointer receiver

### ConversationalAgent
- **Struct size**: 4 fields
- **File**: ../agents/conversational.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `Plan()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver
  - `GetTools()` - pointer receiver
  - `parseOutput()` - pointer receiver

### ConversationalRetrievalQA
- **Struct size**: 9 fields
- **File**: ../chains/conversational_retrieval_qa.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 6
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `getQuestion()` - value receiver
  - `rephraseQuestion()` - value receiver

### CosineDistance
- **Struct size**: 0 fields
- **File**: ../vectorstores/cloudsql/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 8
  - `String()` - value receiver
  - `operator()` - value receiver
  - `searchFunction()` - value receiver
  - `similaritySearchFunction()` - value receiver
  - `String()` - value receiver
  - `operator()` - value receiver
  - `searchFunction()` - value receiver
  - `similaritySearchFunction()` - value receiver

### Cybertron
- **Struct size**: 4 fields
- **File**: ../embeddings/cybertron/cybertron.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `CreateEmbedding()` - pointer receiver

### Definition
- **Struct size**: 6 fields
- **File**: ../jsonschema/json.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `MarshalJSON()` - value receiver

### Documents
- **Struct size**: 2 fields
- **File**: ../tools/metaphor/documents.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `SetOptions()` - pointer receiver
  - `Name()` - pointer receiver
  - `Description()` - pointer receiver
  - `Call()` - pointer receiver
  - `formatContents()` - pointer receiver

### EmbedderClientFunc
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `CreateEmbedding()` - value receiver

### EmbedderImpl
- **Struct size**: 3 fields
- **File**: ../embeddings/embedding.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `EmbedQuery()` - pointer receiver
  - `EmbedDocuments()` - pointer receiver

### EqualsFilter
- **Struct size**: 2 fields
- **File**: ../vectorstores/bedrockknowledgebases/options.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `isFilter()` - value receiver

### Euclidean
- **Struct size**: 0 fields
- **File**: ../vectorstores/cloudsql/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 8
  - `String()` - value receiver
  - `operator()` - value receiver
  - `searchFunction()` - value receiver
  - `similaritySearchFunction()` - value receiver
  - `String()` - value receiver
  - `operator()` - value receiver
  - `searchFunction()` - value receiver
  - `similaritySearchFunction()` - value receiver

### Executor
- **Struct size**: 6 fields
- **File**: ../agents/executor.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 8
  - `Call()` - pointer receiver
  - `doIteration()` - pointer receiver
  - `doAction()` - pointer receiver
  - `getReturn()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver
  - `GetMemory()` - pointer receiver
  - `GetCallbackHandler()` - pointer receiver

### FewShotPrompt
- **Struct size**: 10 fields
- **File**: ../prompts/few_shot.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `getExamples()` - pointer receiver
  - `Format()` - pointer receiver
  - `AssemblePieces()` - pointer receiver
  - `FormatPrompt()` - pointer receiver
  - `GetInputVariables()` - pointer receiver

### FinishReason
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `MarshalJSON()` - value receiver

### FunctionChatMessage
- **Struct size**: 2 fields
- **File**: ../llms/chat_messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `GetType()` - value receiver
  - `GetContent()` - value receiver
  - `GetName()` - value receiver

### GenerateResponse
- **Struct size**: 11 fields
- **File**: ../llms/ollama/internal/ollamaclient/types.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `Summary()` - pointer receiver

### GenericChatMessage
- **Struct size**: 3 fields
- **File**: ../llms/chat_messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `GetType()` - value receiver
  - `GetContent()` - value receiver
  - `GetName()` - value receiver

### GenericMessagePromptTemplate
- **Struct size**: 2 fields
- **File**: ../prompts/message_prompt_template.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `FormatMessages()` - value receiver
  - `GetInputVariables()` - value receiver

### GoogleAI
- **Struct size**: 3 fields
- **File**: ../llms/googleai/new.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 3
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver

### HNSWOptions
- **Struct size**: 2 fields
- **File**: ../vectorstores/cloudsql/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `Options()` - value receiver
  - `Options()` - value receiver

### HTML
- **Struct size**: 1 fields
- **File**: ../documentloaders/html.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `Load()` - value receiver
  - `LoadAndSplit()` - value receiver

### Huggingface
- **Struct size**: 5 fields
- **File**: ../embeddings/huggingface/huggingface.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `EmbedDocuments()` - pointer receiver
  - `EmbedQuery()` - pointer receiver

### HumanChatMessage
- **Struct size**: 1 fields
- **File**: ../llms/chat_messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `GetType()` - value receiver
  - `GetContent()` - value receiver

### HumanMessagePromptTemplate
- **Struct size**: 1 fields
- **File**: ../prompts/message_prompt_template.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `FormatMessages()` - value receiver
  - `GetInputVariables()` - value receiver

### IVFFlatOptions
- **Struct size**: 1 fields
- **File**: ../vectorstores/cloudsql/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `Options()` - value receiver
  - `Options()` - value receiver

### IVFOptions
- **Struct size**: 2 fields
- **File**: ../vectorstores/alloydb/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `Options()` - value receiver

### ImageContent
- **Struct size**: 2 fields
- **File**: ../llms/anthropic/internal/anthropicclient/messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `GetType()` - value receiver

### ImageURLContent
- **Struct size**: 2 fields
- **File**: ../llms/generatecontent.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 4
  - `String()` - value receiver
  - `isPart()` - value receiver
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### InMemory
- **Struct size**: 2 fields
- **File**: ../llms/cache/inmemory/inmemory.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `Get()` - pointer receiver
  - `Put()` - pointer receiver

### IndexSchema
- **Struct size**: 4 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 2
  - `MetadataKeys()` - pointer receiver
  - `AsCommand()` - pointer receiver

### IndexVectorSearch
- **Struct size**: 8 fields
- **File**: ../vectorstores/redisvector/index_search.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `AsCommand()` - value receiver

### InnerProduct
- **Struct size**: 0 fields
- **File**: ../vectorstores/cloudsql/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 8
  - `String()` - value receiver
  - `operator()` - value receiver
  - `searchFunction()` - value receiver
  - `similaritySearchFunction()` - value receiver
  - `String()` - value receiver
  - `operator()` - value receiver
  - `searchFunction()` - value receiver
  - `similaritySearchFunction()` - value receiver

### Jina
- **Struct size**: 6 fields
- **File**: ../embeddings/jina/jina.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 3
  - `EmbedDocuments()` - pointer receiver
  - `EmbedQuery()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver

### KnowledgeBase
- **Struct size**: 4 fields
- **File**: ../vectorstores/bedrockknowledgebases/bedrockknowledgebases.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 20
  - `AddDocuments()` - pointer receiver
  - `AddNamedDocuments()` - pointer receiver
  - `filterMetadata()` - pointer receiver
  - `addDocuments()` - pointer receiver
  - `SimilaritySearch()` - pointer receiver
  - `getFilters()` - pointer receiver
  - `parseMetadata()` - pointer receiver
  - `unmarshalMetadataValue()` - pointer receiver
  - `hash()` - pointer receiver
  - `checkKnowledgeBase()` - pointer receiver
  - `listDataSources()` - pointer receiver
  - `ingestDocuments()` - pointer receiver
  - `startIngestionJob()` - pointer receiver
  - `checkIngestionJobStatus()` - pointer receiver
  - `getOptions()` - pointer receiver
  - `getBucketName()` - pointer receiver
  - `addToS3()` - pointer receiver
  - `removeFromS3()` - pointer receiver
  - `uploadS3Object()` - pointer receiver
  - `removeS3Object()` - pointer receiver

### LLM
- **Struct size**: 3 fields
- **File**: ../llms/watsonx/watsonxllm.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 39
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `getModelPath()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Call()` - pointer receiver
  - `Reset()` - pointer receiver
  - `AddResponse()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `appendGlobalsToArgs()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver

### LLMChain
- **Struct size**: 6 fields
- **File**: ../chains/llm.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 5
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetCallbackHandler()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver

### LLMMathChain
- **Struct size**: 1 fields
- **File**: ../chains/llm_math.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 6
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `processLLMResult()` - value receiver
  - `evaluateExpression()` - value receiver

### LinksSearch
- **Struct size**: 2 fields
- **File**: ../tools/metaphor/links.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `SetOptions()` - pointer receiver
  - `Name()` - pointer receiver
  - `Description()` - pointer receiver
  - `Call()` - pointer receiver
  - `formatLinks()` - pointer receiver

### LogHandler
- **Struct size**: 0 fields
- **File**: ../callbacks/log.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 16
  - `HandleLLMGenerateContentStart()` - value receiver
  - `HandleLLMGenerateContentEnd()` - value receiver
  - `HandleStreamingFunc()` - value receiver
  - `HandleText()` - value receiver
  - `HandleLLMStart()` - value receiver
  - `HandleLLMError()` - value receiver
  - `HandleChainStart()` - value receiver
  - `HandleChainEnd()` - value receiver
  - `HandleChainError()` - value receiver
  - `HandleToolStart()` - value receiver
  - `HandleToolEnd()` - value receiver
  - `HandleToolError()` - value receiver
  - `HandleAgentAction()` - value receiver
  - `HandleAgentFinish()` - value receiver
  - `HandleRetrieverStart()` - value receiver
  - `HandleRetrieverEnd()` - value receiver

### MapReduceDocuments
- **Struct size**: 8 fields
- **File**: ../chains/map_reduce.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 9
  - `Call()` - value receiver
  - `getInputVariable()` - value receiver
  - `maybeAddIntermediateSteps()` - value receiver
  - `getApplyInputs()` - value receiver
  - `mapResultsToReduceInputs()` - value receiver
  - `copyInputValuesWithoutInputKey()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `GetMemory()` - value receiver

### MapRerankDocuments
- **Struct size**: 9 fields
- **File**: ../chains/map_rerank_documents.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 9
  - `Call()` - value receiver
  - `getInputVariable()` - value receiver
  - `getApplyInputs()` - value receiver
  - `copyInputValuesWithoutInputKey()` - value receiver
  - `parseMapResults()` - value receiver
  - `formatOutputs()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `GetMemory()` - value receiver

### MarkdownTextSplitter
- **Struct size**: 8 fields
- **File**: ../textsplitter/markdown_splitter.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `SplitText()` - value receiver

### Memory
- **Struct size**: 10 fields
- **File**: ../memory/zep/zep_memory.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `MemoryVariables()` - pointer receiver
  - `LoadMemoryVariables()` - pointer receiver
  - `SaveContext()` - pointer receiver
  - `Clear()` - pointer receiver
  - `GetMemoryKey()` - pointer receiver

### MessageContent
- **Struct size**: 2 fields
- **File**: ../llms/generatecontent.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 2
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### MessageResponsePayload
- **Struct size**: 8 fields
- **File**: ../llms/anthropic/internal/anthropicclient/messages.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `UnmarshalJSON()` - pointer receiver

### MessagesPlaceholder
- **Struct size**: 1 fields
- **File**: ../prompts/message_prompt_template.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `FormatMessages()` - value receiver
  - `GetInputVariables()` - value receiver

### Model
- **Struct size**: 3 fields
- **File**: ../llms/mistral/mistralmodel.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 3
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver

### MySQL
- **Struct size**: 1 fields
- **File**: ../tools/sqldatabase/mysql/mysql.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `Dialect()` - value receiver
  - `Query()` - value receiver
  - `TableNames()` - value receiver
  - `TableInfo()` - value receiver
  - `Close()` - value receiver

### NoCredentialsError
- **Struct size**: 0 fields
- **File**: ../tools/zapier/internal/errors.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `Error()` - value receiver

### NotEqualsFilter
- **Struct size**: 2 fields
- **File**: ../vectorstores/bedrockknowledgebases/options.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `isFilter()` - value receiver

### NotionDirectoryLoader
- **Struct size**: 2 fields
- **File**: ../documentloaders/notion.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `Load()` - pointer receiver

### NumericField
- **Struct size**: 4 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `AsCommand()` - value receiver

### OneShotZeroAgent
- **Struct size**: 4 fields
- **File**: ../agents/mrkl.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `Plan()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver
  - `GetTools()` - pointer receiver
  - `parseOutput()` - pointer receiver

### OpenAIFunctionsAgent
- **Struct size**: 5 fields
- **File**: ../agents/openai_functions_agent.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 7
  - `functions()` - pointer receiver
  - `Plan()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver
  - `GetTools()` - pointer receiver
  - `constructScratchPad()` - pointer receiver
  - `ParseOutput()` - pointer receiver

### OpenAIOption
- **Struct size**: 0 fields
- **File**: ../agents/options.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `WithSystemMessage()` - value receiver
  - `WithExtraMessages()` - value receiver

### Option
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `apply()` - value receiver

### Options
- **Struct size**: 5 fields
- **File**: ../vectorstores/options.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 3
  - `getMrklPrompt()` - value receiver
  - `getConversationalPrompt()` - value receiver
  - `EnsureAuthPresent()` - pointer receiver

### PDF
- **Struct size**: 3 fields
- **File**: ../documentloaders/pdf.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 3
  - `getPassword()` - pointer receiver
  - `Load()` - value receiver
  - `LoadAndSplit()` - value receiver

### PaLMClient
- **Struct size**: 2 fields
- **File**: ../llms/googleai/internal/palmclient/palmclient.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 6
  - `CreateCompletion()` - pointer receiver
  - `CreateEmbedding()` - pointer receiver
  - `CreateChat()` - pointer receiver
  - `batchPredict()` - pointer receiver
  - `chat()` - pointer receiver
  - `projectLocationPublisherModelPath()` - pointer receiver

### ParseError
- **Struct size**: 2 fields
- **File**: ../outputparser/structured.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `Error()` - value receiver

### PostgreSQL
- **Struct size**: 1 fields
- **File**: ../tools/sqldatabase/postgresql/postgresql.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `Dialect()` - value receiver
  - `Query()` - value receiver
  - `TableNames()` - value receiver
  - `TableInfo()` - value receiver
  - `Close()` - value receiver

### PostgresEngine
- **Struct size**: 1 fields
- **File**: ../util/cloudsqlutil/engine.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 6
  - `Close()` - pointer receiver
  - `InitVectorstoreTable()` - pointer receiver
  - `InitChatHistoryTable()` - pointer receiver
  - `Close()` - pointer receiver
  - `InitVectorstoreTable()` - pointer receiver
  - `InitChatHistoryTable()` - pointer receiver

### PromptTemplate
- **Struct size**: 5 fields
- **File**: ../prompts/prompt_template.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 3
  - `Format()` - value receiver
  - `FormatPrompt()` - value receiver
  - `GetInputVariables()` - value receiver

### RecursiveCharacter
- **Struct size**: 5 fields
- **File**: ../textsplitter/recursive_character.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 3
  - `SplitText()` - value receiver
  - `addSeparatorInSplits()` - value receiver
  - `splitText()` - value receiver

### RedisIndex
- **Struct size**: 4 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `AsCommand()` - pointer receiver

### RefineDocuments
- **Struct size**: 7 fields
- **File**: ../chains/refine_documents.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 7
  - `Call()` - value receiver
  - `constructInitialInputs()` - value receiver
  - `constructRefineInputs()` - value receiver
  - `getBaseInputs()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `GetMemory()` - value receiver

### RegexDict
- **Struct size**: 2 fields
- **File**: ../outputparser/regex_dict.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `GetFormatInstructions()` - value receiver
  - `parse()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `Type()` - value receiver

### RegexParser
- **Struct size**: 2 fields
- **File**: ../outputparser/regex_parser.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `GetFormatInstructions()` - value receiver
  - `parse()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `Type()` - value receiver

### RetrievalQA
- **Struct size**: 4 fields
- **File**: ../chains/retrieval_qa.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 4
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver

### Retriever
- **Struct size**: 4 fields
- **File**: ../vectorstores/vectorstores.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `GetRelevantDocuments()` - value receiver

### RueidisClient
- **Struct size**: 1 fields
- **File**: ../vectorstores/redisvector/redis_client.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 7
  - `DropIndex()` - value receiver
  - `CheckIndexExists()` - value receiver
  - `CreateIndexIfNotExists()` - value receiver
  - `AddDocWithHash()` - value receiver
  - `AddDocsWithHash()` - value receiver
  - `Search()` - value receiver
  - `generateHSetCMD()` - value receiver

### SCANNOptions
- **Struct size**: 2 fields
- **File**: ../vectorstores/alloydb/distance_strategy.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `Options()` - value receiver

### SQLDatabase
- **Struct size**: 3 fields
- **File**: ../tools/sqldatabase/sql_database.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 6
  - `Dialect()` - pointer receiver
  - `TableNames()` - pointer receiver
  - `TableInfo()` - pointer receiver
  - `Query()` - pointer receiver
  - `Close()` - pointer receiver
  - `sampleRows()` - pointer receiver

### SQLDatabaseChain
- **Struct size**: 4 fields
- **File**: ../chains/sql_database.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 4
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver

### SQLite3
- **Struct size**: 1 fields
- **File**: ../tools/sqldatabase/sqlite3/sqlite3.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `Dialect()` - value receiver
  - `Query()` - value receiver
  - `TableNames()` - value receiver
  - `TableInfo()` - value receiver
  - `Close()` - value receiver

### Scraper
- **Struct size**: 5 fields
- **File**: ../tools/scraper/scraper.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 3
  - `Name()` - value receiver
  - `Description()` - value receiver
  - `Call()` - value receiver

### Search
- **Struct size**: 2 fields
- **File**: ../tools/metaphor/search.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `SetOptions()` - pointer receiver
  - `Name()` - pointer receiver
  - `Description()` - pointer receiver
  - `Call()` - pointer receiver
  - `formatResults()` - pointer receiver

### SequentialChain
- **Struct size**: 4 fields
- **File**: ../chains/sequential.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `validateSeqChain()` - pointer receiver
  - `Call()` - pointer receiver
  - `GetMemory()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver

### Simple
- **Struct size**: 0 fields
- **File**: ../outputparser/simple.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 9
  - `MemoryVariables()` - value receiver
  - `LoadMemoryVariables()` - value receiver
  - `SaveContext()` - value receiver
  - `Clear()` - value receiver
  - `GetMemoryKey()` - value receiver
  - `GetFormatInstructions()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `Type()` - value receiver

### SimpleHandler
- **Struct size**: 0 fields
- **File**: ../callbacks/simple.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 16
  - `HandleText()` - value receiver
  - `HandleLLMStart()` - value receiver
  - `HandleLLMGenerateContentStart()` - value receiver
  - `HandleLLMGenerateContentEnd()` - value receiver
  - `HandleLLMError()` - value receiver
  - `HandleChainStart()` - value receiver
  - `HandleChainEnd()` - value receiver
  - `HandleChainError()` - value receiver
  - `HandleToolStart()` - value receiver
  - `HandleToolEnd()` - value receiver
  - `HandleToolError()` - value receiver
  - `HandleAgentAction()` - value receiver
  - `HandleAgentFinish()` - value receiver
  - `HandleRetrieverStart()` - value receiver
  - `HandleRetrieverEnd()` - value receiver
  - `HandleStreamingFunc()` - value receiver

### SimpleSequentialChain
- **Struct size**: 2 fields
- **File**: ../chains/sequential.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 4
  - `Call()` - pointer receiver
  - `GetMemory()` - pointer receiver
  - `GetInputKeys()` - pointer receiver
  - `GetOutputKeys()` - pointer receiver

### SqliteChatMessageHistory
- **Struct size**: 8 fields
- **File**: ../memory/sqlite3/sqlite3_history.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 7
  - `Messages()` - pointer receiver
  - `addMessage()` - pointer receiver
  - `AddMessage()` - pointer receiver
  - `AddAIMessage()` - pointer receiver
  - `AddUserMessage()` - pointer receiver
  - `Clear()` - pointer receiver
  - `SetMessages()` - pointer receiver

### StatusError
- **Struct size**: 3 fields
- **File**: ../llms/ollama/internal/ollamaclient/types.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `Error()` - value receiver
  - `Error()` - value receiver
  - `Error()` - value receiver

### Store
- **Struct size**: 13 fields
- **File**: ../vectorstores/weaviate/weaviate.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 90
  - `AddDocuments()` - pointer receiver
  - `SimilaritySearch()` - pointer receiver
  - `UploadDocument()` - pointer receiver
  - `UploadDocumentAPIRequest()` - pointer receiver
  - `SearchDocuments()` - pointer receiver
  - `httpDefaultSend()` - pointer receiver
  - `CreateIndex()` - pointer receiver
  - `CreateIndexAPIRequest()` - pointer receiver
  - `DeleteIndex()` - pointer receiver
  - `ListIndexes()` - pointer receiver
  - `RetrieveIndex()` - pointer receiver
  - `getOptions()` - pointer receiver
  - `AddDocuments()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `RemoveCollection()` - value receiver
  - `getOptions()` - value receiver
  - `getScoreThreshold()` - value receiver
  - `getNameSpace()` - value receiver
  - `getNamespacedFilter()` - value receiver
  - `init()` - pointer receiver
  - `dropCollection()` - pointer receiver
  - `extractFields()` - pointer receiver
  - `createCollection()` - pointer receiver
  - `createIndex()` - pointer receiver
  - `createSearchParams()` - pointer receiver
  - `getIndex()` - pointer receiver
  - `load()` - pointer receiver
  - `AddDocuments()` - value receiver
  - `getSearchFields()` - pointer receiver
  - `getOptions()` - value receiver
  - `convertResultToDocument()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `getFilters()` - value receiver
  - `AddDocuments()` - pointer receiver
  - `SimilaritySearch()` - pointer receiver
  - `documentIndexing()` - pointer receiver
  - `CreateIndex()` - pointer receiver
  - `DeleteIndex()` - pointer receiver
  - `AddDocuments()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `getOptions()` - value receiver
  - `Close()` - value receiver
  - `init()` - pointer receiver
  - `createVectorExtensionIfNotExists()` - value receiver
  - `createCollectionTableIfNotExists()` - value receiver
  - `createEmbeddingTableIfNotExists()` - value receiver
  - `AddDocuments()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `Search()` - value receiver
  - `DropTables()` - value receiver
  - `RemoveCollection()` - value receiver
  - `createOrGetCollection()` - pointer receiver
  - `getOptions()` - value receiver
  - `getNameSpace()` - value receiver
  - `getScoreThreshold()` - value receiver
  - `getFilters()` - value receiver
  - `deduplicate()` - value receiver
  - `AddDocuments()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `getDocumentsFromMatches()` - value receiver
  - `getNameSpace()` - value receiver
  - `getScoreThreshold()` - value receiver
  - `getFilters()` - value receiver
  - `getOptions()` - value receiver
  - `createProtoStructFilter()` - value receiver
  - `AddDocuments()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `getScoreThreshold()` - value receiver
  - `getFilters()` - value receiver
  - `getOptions()` - value receiver
  - `upsertPoints()` - value receiver
  - `searchPoints()` - value receiver
  - `AddDocuments()` - pointer receiver
  - `SimilaritySearch()` - pointer receiver
  - `DropIndex()` - pointer receiver
  - `getOptions()` - value receiver
  - `getScoreThreshold()` - value receiver
  - `getFilters()` - value receiver
  - `appendDocumentsWithVectors()` - value receiver
  - `AddDocuments()` - value receiver
  - `SimilaritySearch()` - value receiver
  - `MetadataSearch()` - value receiver
  - `parseDocumentsByGraphQLResponse()` - value receiver
  - `deduplicate()` - value receiver
  - `getNameSpace()` - value receiver
  - `getScoreThreshold()` - value receiver
  - `getFilters()` - value receiver
  - `getOptions()` - value receiver
  - `createWhereBuilder()` - value receiver
  - `createFields()` - value receiver

### StreamLogHandler
- **Struct size**: 0 fields
- **File**: ../callbacks/log_stream.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `HandleStreamingFunc()` - value receiver

### StringPromptValue
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `String()` - value receiver
  - `Messages()` - value receiver

### Structured
- **Struct size**: 1 fields
- **File**: ../outputparser/structured.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 5
  - `parse()` - value receiver
  - `Parse()` - value receiver
  - `ParseWithPrompt()` - value receiver
  - `GetFormatInstructions()` - value receiver
  - `Type()` - value receiver

### StuffDocuments
- **Struct size**: 4 fields
- **File**: ../chains/stuff_documents.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 5
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver
  - `joinDocuments()` - value receiver

### SystemChatMessage
- **Struct size**: 1 fields
- **File**: ../llms/chat_messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `GetType()` - value receiver
  - `GetContent()` - value receiver

### SystemMessagePromptTemplate
- **Struct size**: 1 fields
- **File**: ../prompts/message_prompt_template.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `FormatMessages()` - value receiver
  - `GetInputVariables()` - value receiver

### TagField
- **Struct size**: 6 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `AsCommand()` - value receiver

### Text
- **Struct size**: 1 fields
- **File**: ../documentloaders/text.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `Load()` - value receiver
  - `LoadAndSplit()` - value receiver

### TextContent
- **Struct size**: 1 fields
- **File**: ../llms/generatecontent.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 5
  - `GetType()` - value receiver
  - `String()` - value receiver
  - `isPart()` - value receiver
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### TextField
- **Struct size**: 8 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `AsCommand()` - value receiver

### TokenSplitter
- **Struct size**: 6 fields
- **File**: ../textsplitter/token_splitter.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 2
  - `SplitText()` - value receiver
  - `splitText()` - value receiver

### Tool
- **Struct size**: 6 fields
- **File**: ../tools/zapier/zapier.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 17
  - `Name()` - value receiver
  - `Description()` - value receiver
  - `Call()` - value receiver
  - `Name()` - pointer receiver
  - `Description()` - pointer receiver
  - `Call()` - pointer receiver
  - `Name()` - value receiver
  - `Description()` - value receiver
  - `Call()` - value receiver
  - `Name()` - value receiver
  - `Description()` - value receiver
  - `Call()` - value receiver
  - `searchWiKi()` - value receiver
  - `Name()` - value receiver
  - `Description()` - value receiver
  - `Call()` - value receiver
  - `createDescription()` - value receiver

### ToolCall
- **Struct size**: 3 fields
- **File**: ../llms/openai/internal/openaiclient/chat.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 3
  - `isPart()` - value receiver
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### ToolCallResponse
- **Struct size**: 3 fields
- **File**: ../llms/generatecontent.go
- **⚠️  MIXED RECEIVERS** - This is problematic!
- **Methods**: 3
  - `isPart()` - value receiver
  - `MarshalJSON()` - value receiver
  - `UnmarshalJSON()` - pointer receiver

### ToolChatMessage
- **Struct size**: 2 fields
- **File**: ../llms/chat_messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `GetType()` - value receiver
  - `GetContent()` - value receiver
  - `GetID()` - value receiver

### ToolOptions
- **Struct size**: 7 fields
- **File**: ../tools/zapier/zapier.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `Validate()` - value receiver

### ToolResultContent
- **Struct size**: 3 fields
- **File**: ../llms/anthropic/internal/anthropicclient/messages.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 1
  - `GetType()` - value receiver

### ToolUseContent
- **Struct size**: 4 fields
- **File**: ../llms/anthropic/internal/anthropicclient/messages.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `GetType()` - value receiver

### Transform
- **Struct size**: 4 fields
- **File**: ../chains/transform.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 4
  - `Call()` - value receiver
  - `GetMemory()` - value receiver
  - `GetInputKeys()` - value receiver
  - `GetOutputKeys()` - value receiver

### Transport
- **Struct size**: 4 fields
- **File**: ../tools/zapier/internal/client.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 3
  - `RoundTrip()` - pointer receiver
  - `createAuthHeader()` - pointer receiver
  - `createHeaders()` - pointer receiver

### VectorField
- **Struct size**: 12 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **⚠️  Value receivers on large struct** - Consider pointer receivers
- **Methods**: 1
  - `AsCommand()` - value receiver

### VectorStore
- **Struct size**: 11 fields
- **File**: ../vectorstores/cloudsql/vectorstore.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 22
  - `AddDocuments()` - pointer receiver
  - `generateAddDocumentsQuery()` - pointer receiver
  - `SimilaritySearch()` - pointer receiver
  - `executeSQLQuery()` - pointer receiver
  - `processResultsToDocuments()` - pointer receiver
  - `ApplyVectorIndex()` - pointer receiver
  - `ReIndex()` - pointer receiver
  - `ReIndexWithName()` - pointer receiver
  - `DropVectorIndex()` - pointer receiver
  - `IsValidIndex()` - pointer receiver
  - `NewBaseIndex()` - pointer receiver
  - `AddDocuments()` - pointer receiver
  - `generateAddDocumentsQuery()` - pointer receiver
  - `SimilaritySearch()` - pointer receiver
  - `executeSQLQuery()` - pointer receiver
  - `processResultsToDocuments()` - pointer receiver
  - `ApplyVectorIndex()` - pointer receiver
  - `ReIndex()` - pointer receiver
  - `ReIndexWithName()` - pointer receiver
  - `DropVectorIndex()` - pointer receiver
  - `IsValidIndex()` - pointer receiver
  - `NewBaseIndex()` - pointer receiver

### Vertex
- **Struct size**: 4 fields
- **File**: ../llms/googleai/vertex/new.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 3
  - `CreateEmbedding()` - pointer receiver
  - `Call()` - pointer receiver
  - `GenerateContent()` - pointer receiver

### VoyageAI
- **Struct size**: 6 fields
- **File**: ../embeddings/voyageai/voyageai.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 4
  - `EmbedDocuments()` - pointer receiver
  - `EmbedQuery()` - pointer receiver
  - `request()` - pointer receiver
  - `decodeError()` - pointer receiver

### chromaGoEmbedder
- **Struct size**: 0 fields
- **File**: ../vectorstores/chroma/embedder.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `EmbedDocuments()` - value receiver
  - `EmbedQuery()` - value receiver
  - `EmbedRecords()` - value receiver

### ingestDocumentsRetryer
- **Struct size**: 0 fields
- **File**: ../vectorstores/bedrockknowledgebases/ingestion.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 3
  - `hasReachedMaxNumberRequests()` - value receiver
  - `hasReachedMaxConcurrency()` - value receiver
  - `IsErrorRetryable()` - value receiver

### logJSONTransport
- **Struct size**: 1 fields
- **File**: ../httputil/debug_transport_body.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `RoundTrip()` - pointer receiver

### logTransport
- **Struct size**: 1 fields
- **File**: ../httputil/debug_transport.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `RoundTrip()` - pointer receiver

### markdownContext
- **Struct size**: 20 fields
- **File**: ../textsplitter/markdown_splitter.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 22
  - `splitText()` - pointer receiver
  - `clone()` - pointer receiver
  - `onMDHeader()` - pointer receiver
  - `onMDParagraph()` - pointer receiver
  - `onMDQuote()` - pointer receiver
  - `onMDBulletList()` - pointer receiver
  - `onMDOrderedList()` - pointer receiver
  - `onMDList()` - pointer receiver
  - `onMDListItem()` - pointer receiver
  - `onMDListItemParagraph()` - pointer receiver
  - `onMDTable()` - pointer receiver
  - `splitTableRows()` - pointer receiver
  - `onTableHeader()` - pointer receiver
  - `onTableBody()` - pointer receiver
  - `onMDCodeBlock()` - pointer receiver
  - `onMDFence()` - pointer receiver
  - `onMDHr()` - pointer receiver
  - `joinSnippet()` - pointer receiver
  - `applyToChunks()` - pointer receiver
  - `splitInline()` - pointer receiver
  - `inlineOnLinkClose()` - pointer receiver
  - `inlineOnImage()` - pointer receiver

### mockEmbedder
- **Struct size**: 3 fields
- **File**: ../vectorstores/mongovector/mock_embedder.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 4
  - `mockDocuments()` - pointer receiver
  - `existingVectors()` - pointer receiver
  - `EmbedDocuments()` - pointer receiver
  - `EmbedQuery()` - pointer receiver

### mockLLM
- **Struct size**: 2 fields
- **File**: ../vectorstores/mongovector/mock_llm.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `CreateEmbedding()` - pointer receiver

### parser
- **Struct size**: 4 fields
- **File**: ../prompts/internal/fstring/parser.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 5
  - `parse()` - pointer receiver
  - `scanToLeftCurlyBracket()` - pointer receiver
  - `scanToRightCurlyBracket()` - pointer receiver
  - `hasMore()` - pointer receiver
  - `get()` - pointer receiver

### schemaGenerator
- **Struct size**: 3 fields
- **File**: ../vectorstores/redisvector/index_schema.go
- **✅ Pointer receivers** - Good for larger structs or when mutation is needed
- **Methods**: 1
  - `generate()` - pointer receiver

### startIngestionJobRetryer
- **Struct size**: 0 fields
- **File**: ../vectorstores/bedrockknowledgebases/ingestion.go
- **✅ Value receivers** - Appropriate for small struct
- **Methods**: 2
  - `hasOngoingIngestDocsRequest()` - value receiver
  - `IsErrorRetryable()` - value receiver

## Recommendations

### 🚨 Issues Found

The following types should be reviewed:
#### AIChatMessage
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### BinaryContent
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### ChatMessage
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### ConversationalRetrievalQA
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 9 fields may be expensive to copy

#### Definition
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 6 fields may be expensive to copy

#### ImageURLContent
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### IndexVectorSearch
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 8 fields may be expensive to copy

#### LLMChain
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 6 fields may be expensive to copy

#### MapReduceDocuments
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 8 fields may be expensive to copy

#### MapRerankDocuments
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 9 fields may be expensive to copy

#### MarkdownTextSplitter
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 8 fields may be expensive to copy

#### MessageContent
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### NumericField
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### Options
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### PDF
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### PromptTemplate
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 5 fields may be expensive to copy

#### RecursiveCharacter
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 5 fields may be expensive to copy

#### RefineDocuments
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 7 fields may be expensive to copy

#### RetrievalQA
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### Retriever
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### SQLDatabaseChain
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### Scraper
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 5 fields may be expensive to copy

#### Store
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### StuffDocuments
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### TagField
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 6 fields may be expensive to copy

#### TextContent
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### TextField
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 8 fields may be expensive to copy

#### TokenSplitter
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 6 fields may be expensive to copy

#### Tool
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### ToolCall
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### ToolCallResponse
**Issue**: Mixed receiver types
**Recommendation**: Choose either all value or all pointer receivers
**Guideline**: Use pointer receivers if:
- The struct is large (>3-4 fields)
- Methods need to modify the receiver
- The struct contains sync.Mutex or similar fields
- Other methods already use pointer receivers

#### ToolOptions
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 7 fields may be expensive to copy

#### ToolUseContent
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### Transform
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 4 fields may be expensive to copy

#### VectorField
**Issue**: Value receivers on potentially large struct
**Recommendation**: Consider switching to pointer receivers to avoid copying
**Details**: 12 fields may be expensive to copy

### General Guidelines

**Use pointer receivers when:**
- The method needs to modify the receiver
- The struct is large (typically >3-4 fields)
- The struct contains sync.Mutex or similar types
- To maintain consistency if some methods already use pointer receivers

**Use value receivers when:**
- The struct is small (1-3 simple fields)
- The method doesn't modify the receiver
- You want to ensure the receiver is immutable
- The type is a basic type like int, string, or small struct
