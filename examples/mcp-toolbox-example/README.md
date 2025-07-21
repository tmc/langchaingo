# 🧰 MCP Toolbox for Databases Example

[MCP Toolbox for Databases](https://github.com/googleapis/genai-toolbox) is an open source MCP server for databases. It was designed with enterprise-grade and production-quality in mind. It enables you to develop tools easier, faster, and more securely by handling the complexities such as connection pooling, authentication, and more.
For more information on how to get started, see the [documentation](https://googleapis.github.io/genai-toolbox/getting-started/).
This example demonstrates how to use  Toolbox Tools seemlessly with LangChain Go.

## What does this example do?

The `mcp_toolbox_example.go` file showcases how to:

- Initializes a client for the MCP Toolbox server
- Fetches tools from the server
- Converts into LangChain Go compatible tools

## How to use this example

Before you get started:

1. Install the MCP Toolbox Server:
  Follow the detailed instructions in the [official documentation](https://googleapis.github.io/genai-toolbox/getting-started/introduction/#installing-the-server).

2. Configure Your Toolbox:
  Refer to the [configuration guide](https://googleapis.github.io/genai-toolbox/getting-started/configure/) to set up your toolbox.

3. Set Your Google API Key:
  ```shell
   export GOOGLE_API_KEY=<your_google_api_key>
   ```

4. Run the example:
  ```go
   go run mcp_toolbox_example.go
  ```

## Key Features

Toolbox has a variety of features to make developing Gen AI tools for databases.
For more information, read more about the following:

* [Authenticated Parameters](https://googleapis.github.io/genai-toolbox/resources/tools/#authenticated-parameters): bind tool inputs to values from OIDC tokens automatically, making it easy to run sensitive queries without potentially leaking data
* [Authorized Invocations:](https://googleapis.github.io/genai-toolbox/resources/tools/#authorized-invocations) restrict access to use a tool based on the users Auth token
* [OpenTelemetry](https://googleapis.github.io/genai-toolbox/how-to/export_telemetry/): get metrics and tracing from Toolbox with OpenTelemetry