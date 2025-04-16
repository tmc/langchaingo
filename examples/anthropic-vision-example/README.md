# Anthropic Vision Example

Hello there! 👋 This example demonstrates how to use the Anthropic Claude 3 Claude 3 Sonnet model for image analysis using Go and the LangChain Go library. Let's break down what this exciting code does!

## What This Example Does

1. **Sets Up Anthropic**: The code initializes an Anthropic client to interact with the Claude 3 Sonnet model.

2. **Loads an Image**: An image file (`image.png`) is embedded into the binary using Go's `embed` package. This image will be analyzed by the AI model.

3. **Sends a Request**: The code constructs a request to the Claude 3 model, including:
   - The image data in a base64 encoded string (in PNG format)
   - A text prompt asking to identify the string on a box in the image

4. **Processes the Response**: After sending the request, the code handles the response from the AI model, extracting the generated content and some metadata about token usage.

5. **Outputs Results**: Finally, it prints out the AI's interpretation of what string is on the box in the image.

## Key Features

- **Multimodal AI**: This example showcases how to work with both image and text inputs in a single AI request.
- **Error Handling**: Includes basic error checking to ensure the process runs smoothly.
- **Token Usage Tracking**: Logs the number of input and output tokens used, which can be helpful for monitoring usage and costs.

## Running the Example

To run this example, you'll need:

1. An Anthropic API KEY set up in your environment variables
2. The required Go dependencies installed

Once everything is set up, simply run the Go file, and it should output the AI's interpretation of the text on the box in the image!

Happy coding, and enjoy exploring the fascinating world of multimodal AI with Claude 3! 🚀🖼️🤖
