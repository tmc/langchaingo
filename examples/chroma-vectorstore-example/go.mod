module github.com/tmc/langchaingo/examples/chroma-vectorstore-example

go 1.21

toolchain go1.21.4

require (
	github.com/amikos-tech/chroma-go v0.0.2
	github.com/google/uuid v1.6.0
	github.com/tmc/langchaingo v0.1.4
)

require (
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
)

replace github.com/tmc/langchaingo => ../../
