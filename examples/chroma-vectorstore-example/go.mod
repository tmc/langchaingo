module chroma-vectorstore-example

go 1.20

// NOTE: remove the following line to use the official (rather than local development) version
replace github.com/tmc/langchaingo => ../..

require (
	github.com/google/uuid v1.3.0
	github.com/tmc/langchaingo v0.0.0-20230829032009-e89bc0bd369f
)

require (
	github.com/amikos-tech/chroma-go v0.0.0-20230901221218-d0087270239e // indirect
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.2 // indirect
)
