This directory contains langchaingo provider for Google's models.

* In the main `googleai` directory: provider for Google AI
  (https://ai.google.dev/)
* In the `vertex` directory: provider for GCP Vertex AI
  (https://cloud.google.com/vertex-ai/)
* In the `palm` directory: provider for the legacy PaLM models.

Both the `googleai` and `vertex` providers give access to Gemini-family
multi-modal LLMs. The `vertex` package implements Vertex AI functionality
using `google.golang.org/genai` library.

----

Testing:

The test code between `googleai` and `vertex` is also shared, and lives in
the `shared_test` directory. The same tests are run for both providers.
