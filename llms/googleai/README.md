This directory contains langchaingo provider for Google's models.

* In the main `googleai` directory: provider for Google AI
  (https://ai.google.dev/)
* In the `vertex` directory: provider for GCP Vertex AI
  (https://cloud.google.com/vertex-ai/)
* In the `palm` directory: provider for the legacy PaLM models.

Both the `googleai` and `vertex` providers give access to Gemini-family
multi-modal LLMs. The code between these providers was historically very similar,
and most of the `vertex` package used to be code-generated from the `googleai`
package.

That generator is now **obsolete and must not be run**: the two packages target
different SDKs — `googleai` uses `google.golang.org/genai` while `vertex` uses
`cloud.google.com/go/vertexai/genai` — so their sources have legitimately
diverged. The `vertex` package is now **hand-maintained**. The former generator at
`llms/googleai/internal/cmd/generate-vertex.go` exits with an error explaining
this; migrating `vertex` onto the unified SDK is tracked as separate future work.

----

Testing:

The test code between `googleai` and `vertex` is also shared, and lives in
the `shared_test` directory. The same tests are run for both providers.
