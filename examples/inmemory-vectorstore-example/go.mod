module github.com/tmc/langchaingo/examples/inmemory-vectorstore-example

go 1.21

toolchain go1.21.4

require github.com/tmc/langchaingo v0.1.4-alpha.0

require (
	github.com/Bithack/go-hnsw v0.0.0-20170629124716-52a932462077 // indirect
	github.com/chewxy/math32 v1.10.1 // indirect
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/pkoukk/tiktoken-go v0.1.2 // indirect
	github.com/willf/bitset v1.1.11 // indirect
)

replace github.com/tmc/langchaingo => ../..
