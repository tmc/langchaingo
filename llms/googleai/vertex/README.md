This package implements a langchaingo provider for Google Vertex AI LLMs,
using the SDK at https://pkg.go.dev/cloud.google.com/go/vertexai

Since Vertex SDK is so similar to the Google AI SDK, we generate the main part
of this package from `llms/googleai/googleai.go` to create
`llms/googleai/vertex/vertex.go`.

To re-generate, run this from the root of the repository:

    go run ./llms/googleai/internal/cmd/generate-vertex.go < llms/googleai/googleai.go > llms/googleai/vertex/vertex.go

See the script in `llms/googleai/internal/cmd/generate-vertex.go` for details.
