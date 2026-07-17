# Google AI Structured Output Example

This example shows how to get **schema-constrained structured output** from a
Google Gemini model with the `langchaingo` library. Instead of parsing free-form
text, you pass a JSON Schema and get back a JSON value that matches it.

## What This Example Does

1. Sets up a Google AI (Gemini) client using your API key.
2. Defines a JSON Schema describing the shape it wants back (a Moon landing:
   `mission`, `year`, and an `astronauts` array).
3. Calls `GenerateContent` with `llms.WithStructuredOutput(...)`, passing that schema.
4. Unmarshals the response straight into a typed Go struct and prints it.

## How It Works

- `llms.WithStructuredOutput(llms.StructuredOutputConfig{...})` is the
  provider-neutral way to request structured output. It carries a raw JSON Schema
  (Draft 2020-12); the SDK sends it to Gemini via `responseJsonSchema` and sets the
  response MIME type to `application/json`.
- On a normal completion the SDK **validates the response against your original
  schema** before returning it. So `resp.Choices[0].Content` is guaranteed to be a
  single JSON value matching the schema and unmarshals directly into a struct —
  no prompt engineering or manual cleanup.
- If the model returns something that does not match, you get a typed
  `*llms.ErrStructuredOutputValidation` instead of a silent bad result.

The same `WithStructuredOutput` option works across providers (OpenAI, Anthropic,
Amazon Bedrock, Vertex, Ollama) — only the model and API key change.

## Running the Example

1. Set your Google AI API key:
   ```
   export GOOGLE_API_KEY=your_api_key_here
   ```

2. Run the program:
   ```
   go run googleai-structured-output-example.go
   ```

3. Expected output (values come from the model):
   ```
   Mission:    Apollo 11
   Year:       1969
   Astronauts: [Neil Armstrong Buzz Aldrin Michael Collins]
   ```
