module chroma-vectorstore-example

go 1.20

// NOTE: remove the following line to use the official (rather than local development) version
replace github.com/tmc/langchaingo => ../..

require (
	github.com/amikos-tech/chroma-go v0.0.0-20230901221218-d0087270239e
	github.com/google/uuid v1.3.1
	github.com/tmc/langchaingo v0.0.0-20231020205806-b33244eb8de8
)

require (
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.2 // indirect
	golang.org/x/exp v0.0.0-20230510235704-dd950f8aeaea // indirect
)
