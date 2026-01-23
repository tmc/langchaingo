module github.com/vxcontrol/langchaingo/examples/chroma-vectorstore-example

go 1.24.1

require (
	github.com/amikos-tech/chroma-go v0.2.3
	github.com/google/uuid v1.6.0
	github.com/vxcontrol/langchaingo v0.1.14-update.1
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.8 // indirect
	github.com/yalue/onnxruntime_go v1.19.0 // indirect
)

replace github.com/vxcontrol/langchaingo => ../..
