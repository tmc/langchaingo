module openrouter-llm-example

go 1.23.0

toolchain go1.24.1

require github.com/tmc/langchaingo v0.1.13

replace github.com/tmc/langchaingo => ../..

require (
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
)
