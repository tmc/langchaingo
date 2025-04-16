# ERNIE Function Call Example

This example demonstrates how to use the ERNIE (Enhanced Representation through kNowledge IntEgration) language model with function calling capabilities using the LangChain Go library.

## What This Example Does

1. **Initializes the ERNIE Model**: 
   - Sets up the ERNIE model with specific configurations, including the model name and authentication details.

2. **Defines a Weather Function**:
   - Implements a `getCurrentWeather` function that simulates fetching weather data for a given location.

3. **Registers Function Definition**:
   - Defines a function that can be called by the ERNIE model, specifying its name, description, and parameters.

4. **Generates Content with Function Call**:
   - Sends a query about the weather in Boston to the ERNIE model.
   - The model is provided with the capability to call the `getCurrentWeather` function if needed.

5. **Handles the Response**:
   - Checks if the model's response includes a function call.
   - If a function call is present, it prints the details of the call.

## Key Features

- **ERNIE Model Integration**: Demonstrates how to initialize and use the ERNIE model in a Go application.
- **Function Calling**: Shows how to define and use custom functions that can be called by the language model.
- **Error Handling**: Includes basic error handling for model initialization and content generation.

## Usage Notes

- You need to replace the placeholder values for `ak`, `sk`, and `accesstoken` with your actual ERNIE API credentials.
- This example is designed to showcase the function calling feature. In a real-world scenario, you would implement actual weather data fetching logic in the `getCurrentWeather` function.

## Running the Example

To run this example, ensure you have the necessary dependencies installed and your ERNIE API credentials set up. Then, execute the Go file to see the model's response and any potential function calls it makes.

This example provides a great starting point for developers looking to integrate ERNIE's language capabilities with custom function calling in their Go applications.
