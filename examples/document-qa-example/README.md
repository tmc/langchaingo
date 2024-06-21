# Document QA Example

This example demonstrates how to use the LangChain Go library to perform question answering on documents using two different approaches: Stuff QA and Refine QA.

## What This Example Does

1. It sets up an OpenAI language model (LLM) client.

2. It demonstrates two methods of question answering:

   a) Stuff QA Chain:
   - Creates a chain that takes input documents and a question.
   - Combines all documents into a single prompt for the LLM.
   - Suitable for a small number of documents.
   - Asks: "Where did Harrison go to collage?"

   b) Refine QA Chain:
   - Creates a chain that iterates over input documents one by one.
   - Updates an intermediate answer with each iteration.
   - Uses the previous version of the answer and the next document as context.
   - Requires multiple LLM calls (can't be done in parallel).
   - Asks: "Where did Ankush go to collage?"

3. For each method, it:
   - Prepares sample documents about Harrison and Ankush's education.
   - Calls the respective chain with the documents and question.
   - Prints the answer received from the LLM.

## How to Run

1. Ensure you have Go installed on your system.
2. Set up your OpenAI API key as an environment variable.
3. Run the example using: `go run document_qa.go`

## Note

This example showcases two different approaches to document QA, allowing you to compare their usage and effectiveness for different scenarios. The Stuff QA method is simpler and faster for small document sets, while the Refine QA method can handle larger document sets by processing them iteratively.

Happy questioning! 📚🔍
