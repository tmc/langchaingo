# WatsonX LLM Example

This example demonstrates how to use the WatsonX language model with LangChain Go to generate text completions. It's a simple and fun way to explore the capabilities of large language models!

## What This Example Does

1. **Sets up the WatsonX LLM**: The code initializes a WatsonX language model using the "meta-llama/llama-3-70b-instruct" model. You can easily customize this by uncommenting and providing your own API key and project ID.

2. **Generates a Company Name**: The example asks the LLM to come up with a creative company name for a business that makes colorful socks. It's a playful prompt that showcases the model's ability to generate creative ideas.

3. **Customizes Generation Parameters**: The code demonstrates how to fine-tune the text generation process using parameters like `TopK`, `TopP`, and `Seed`. These allow you to control the creativity and consistency of the output.

4. **Prints the Result**: Finally, the generated company name suggestion is printed to the console.

## How to Run

1. Ensure you have Go installed on your system.
2. Set up your WatsonX credentials (API key and project ID) if needed.
3. Run the example with: `go run watsonx_example.go`

## Fun Twist

Every time you run this example, you might get a different, creative company name for your imaginary colorful sock business. It's like having a brainstorming session with an AI!

Enjoy exploring the world of AI-generated business names and have fun with your virtual sock company! 🧦🌈
