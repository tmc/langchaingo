# Zapier LLM Example with LangChain Go

Hello there! 👋 This example demonstrates how to use LangChain Go to create an AI agent that can interact with Zapier's Natural Language Actions (NLA) API. It's a fantastic way to combine the power of language models with the automation capabilities of Zapier!

## What This Example Does

This example showcases the following:

1. **Setting up an OpenAI Language Model**: It initializes an OpenAI LLM, which will be used as the brain of our agent.

2. **Integrating Zapier NLA Tools**: The code fetches all available Zapier NLA tools using your Zapier API key.

3. **Creating an AI Agent**: It sets up an agent using the "Zero Shot React Description" approach, which allows the agent to decide how to use tools based on their descriptions.

4. **Executing a Task**: The agent is given a specific task: "Get the last email from noreply@github.com". It will use the Zapier tools to accomplish this task.

5. **Displaying Results**: Finally, it prints out the agent's response to the given task.

## How to Use

1. Make sure you have your OpenAI API key set in your environment variables.
2. Set your Zapier NLA API key in the `ZAPIER_NLA_API_KEY` environment variable.
3. Run the example, and watch as the AI agent uses Zapier tools to fetch the last email from noreply@github.com!

This example is a great starting point for building more complex AI-powered automation workflows that leverage both language models and Zapier's vast ecosystem of app integrations.

Happy coding! 🚀
