module github.com/tmc/langchaingo/examples/chroma-vectorstore-example

go 1.23.0

toolchain go1.24.4

require (
	github.com/amikos-tech/chroma-go v0.1.4
	github.com/google/uuid v1.6.0
	github.com/tmc/langchaingo v0.1.13
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
)

replace github.com/tmc/langchaingo => ../..
