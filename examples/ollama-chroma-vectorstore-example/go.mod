module github.com/vendasta/langchaingo/examples/ollama-chroma-vectorstore-example

go 1.24

toolchain go1.24.6

require (
	github.com/google/uuid v1.6.0
	github.com/vendasta/langchaingo v0.1.14-pre.4
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/amikos-tech/chroma-go v0.1.4 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.8 // indirect
)

replace github.com/vendasta/langchaingo => ../..
