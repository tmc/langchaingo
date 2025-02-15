This directory contains langchaingo provider for Google's models.

* In the main `googlegenai` directory: provider for Google AI, supports (Vertex & Gemini)
  (https://ai.google.dev/)
* In the main `googleai` directory: provider for LEGACY Google AI
  (https://ai.google.dev/)
* In the `vertex` directory: provider for GCP LEGACY Vertex AI
  (https://cloud.google.com/vertex-ai/)
* In the `palm` directory: provider for the legacy PaLM models.

Both the `googleai` and `vertex` providers give access to Gemini-family
multi-modal LLMs. The code between these providers is very similar; therefore,
most of the `vertex` package is code-generated from the `googleai` package using
a tool:

    go run ./llms/googleai/internal/cmd/generate-vertex.go < llms/googleai/googleai.go > llms/googleai/vertex/vertex.go

----

Testing:

The test code between `googleai` and `vertex` is also shared, and lives in
the `shared_test` directory. The same tests are run for both providers.
