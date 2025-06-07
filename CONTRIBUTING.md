
# Contributing to langchaingo

First off, thanks for taking the time to contribute! ❤️

All types of contributions are encouraged and valued. See the [Table of Contents](#table-of-contents) for different ways to help and details about how this project handles them. Please make sure to read the relevant section before making your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. The community looks forward to your contributions. 🎉

> And if you like the project, but just don't have time to contribute, that's fine. There are other easy ways to support the project and show your appreciation, which we would also be very happy about:
> - Star the project
> - Tweet about it
> - Refer this project in your project's readme
> - Mention the project at local meetups and tell your friends/colleagues

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [I Have a Question](#i-have-a-question)
- [I Want To Contribute](#i-want-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
    - [Before Submitting a Bug Report](#before-submitting-a-bug-report)
    - [How Do I Submit a Good Bug Report?](#how-do-i-submit-a-good-bug-report)
  - [Suggesting Enhancements](#suggesting-enhancements)
    - [Before Submitting an Enhancement](#before-submitting-an-enhancement)
    - [How Do I Submit a Good Enhancement Suggestion?](#how-do-i-submit-a-good-enhancement-suggestion)
  - [Your First Code Contribution](#your-first-code-contribution)
    - [Make Changes](#make-changes)
      - [Make changes in the UI](#make-changes-in-the-ui)
      - [Make changes locally](#make-changes-locally)
    - [Running Tests](#running-tests)
    - [Testing with httprr](#testing-with-httprr)
      - [How httprr works](#how-httprr-works)
      - [Writing tests with httprr](#writing-tests-with-httprr)
      - [Recording new tests](#recording-new-tests)
      - [Important notes about httprr](#important-notes-about-httprr)
      - [Debugging httprr issues](#debugging-httprr-issues)
    - [Commit your update](#commit-your-update)
    - [Pull Request](#pull-request)
    - [Your PR is merged!](#your-pr-is-merged)

## Code of Conduct

This project and everyone participating in it is governed by the
[langchaingo Code of Conduct](CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code. Please report unacceptable behavior
to <travis.cline@gmail.com>.


## I Have a Question

> If you want to ask a question, we assume that you have read the available [Documentation](https://pkg.go.dev/github.com/tmc/langchaingo).

Before you ask a question, it is best to search for existing [Issues](https://github.com/tmc/langchaingo/issues) that might help you. In case you have found a suitable issue and still need clarification, you can write your question in this issue. It is also advisable to search the internet for answers first.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [Issue](https://github.com/tmc/langchaingo/issues/new).
- Provide as much context as you can about what you're running into.
- Provide project and platform versions (nodejs, npm, etc), depending on what seems relevant.

We will then take care of the issue as soon as possible.

## I Want To Contribute

> ### Legal Notice
> When contributing to this project, you must agree that you have authored 100% of the content, that you have the necessary rights to the content and that the content you contribute may be provided under the project license.

### Reporting Bugs

#### Before Submitting a Bug Report

A good bug report shouldn't leave others needing to chase you up for more information. Therefore, we ask you to investigate carefully, collect information and describe the issue in detail in your report. Please complete the following steps in advance to help us fix any potential bug as fast as possible.

- Make sure that you are using the latest version.
- Determine if your bug is really a bug and not an error on your side e.g. using incompatible environment components/versions (Make sure that you have read the [documentation](https://pkg.go.dev/github.com/tmc/langchaingo). If you are looking for support, you might want to check [this section](#i-have-a-question)).
- To see if other users have experienced (and potentially already solved) the same issue you are having, check if there is not already a bug report existing for your bug or error in the [bug tracker](https://github.com/tmc/langchaingo/issues?q=label%3Abug).
- Also make sure to search the internet (including Stack Overflow) to see if users outside of the GitHub community have discussed the issue.
- Collect information about the bug:
  - Stack trace (Traceback)
  - OS, Platform and Version (Windows, Linux, macOS, x86, ARM)
  - Version of the interpreter, compiler, SDK, runtime environment, package manager, depending on what seems relevant.
  - Possibly your input and the output
  - Can you reliably reproduce the issue? And can you also reproduce it with older versions?

#### How Do I Submit a Good Bug Report?

> You must never report security related issues, vulnerabilities or bugs including sensitive information to the issue tracker, or elsewhere in public. Instead sensitive bugs must be sent by email to <travis.cline@gmail.com>.
<!-- You may add a PGP key to allow the messages to be sent encrypted as well. -->

We use GitHub issues to track bugs and errors. If you run into an issue with the project:

- Open an [Issue](https://github.com/tmc/langchaingo/issues/new). (Since we can't be sure at this point whether it is a bug or not, we ask you not to talk about a bug yet and not to label the issue.)
- Explain the behavior you would expect and the actual behavior.
- Please provide as much context as possible and describe the *reproduction steps* that someone else can follow to recreate the issue on their own. This usually includes your code. For good bug reports you should isolate the problem and create a reduced test case.
- Provide the information you collected in the previous section.

Once it's filed:

- The project team will label the issue accordingly.
- A team member will try to reproduce the issue with your provided steps. If there are no reproduction steps or no obvious way to reproduce the issue, the team will ask you for those steps and mark the issue as `needs-repro`. Bugs with the `needs-repro` tag will not be addressed until they are reproduced.
- If the team is able to reproduce the issue, it will be marked `needs-fix`, as well as possibly other tags (such as `critical`), and the issue will be left to be [implemented by someone](#your-first-code-contribution).

<!-- You might want to create an issue template for bugs and errors that can be used as a guide and that defines the structure of the information to be included. If you do so, reference it here in the description. -->


### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for langchaingo, **including completely new features and minor improvements to existing functionality**. Following these guidelines will help maintainers and the community to understand your suggestion and find related suggestions.

#### Before Submitting an Enhancement

- Make sure that you are using the latest version.
- Read the [documentation](https://pkg.go.dev/github.com/tmc/langchaingo) carefully and find out if the functionality is already covered, maybe by an individual configuration.
- Perform a [search](https://github.com/tmc/langchaingo/issues) to see if the enhancement has already been suggested. If it has, add a comment to the existing issue instead of opening a new one.
- Find out whether your idea fits with the scope and aims of the project. It's up to you to make a strong case to convince the project's developers of the merits of this feature. Keep in mind that we want features that will be useful to the majority of our users and not just a small subset. If you're just targeting a minority of users, consider writing an add-on/plugin library.

#### How Do I Submit a Good Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://github.com/tmc/langchaingo/issues).

- Use a **clear and descriptive title** for the issue to identify the suggestion.
- Provide a **step-by-step description of the suggested enhancement** in as many details as possible.
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why. At this point you can also tell which alternatives do not work for you.
- You may want to **include screenshots and animated GIFs** which help you demonstrate the steps or point out the part which the suggestion is related to. You can use [this tool](https://www.cockos.com/licecap/) to record GIFs on macOS and Windows, and [this tool](https://github.com/colinkeenan/silentcast) or [this tool](https://github.com/GNOME/byzanz) on Linux. <!-- this should only be included if the project has a GUI -->
- **Explain why this enhancement would be useful** to most langchaingo users. You may also want to point out the other projects that solved it better and which could serve as inspiration.
- We strive to conceptually align with the Python and TypeScript versions of Langchain. Please link/reference the associated concepts in those codebases when introducing a new concept.

<!-- You might want to create an issue template for enhancement suggestions that can be used as a guide and that defines the structure of the information to be included. If you do so, reference it here in the description. -->

### Your First Code Contribution

#### Make Changes

##### Make changes in the UI

Click **Make a contribution** at the bottom of any docs page to make small changes such as a typo, sentence fix, or a broken link. This takes you to the `.md` file where you can make your changes and [create a pull request](#pull-request) for a review.

##### Make changes locally

1. Fork the repository.
- Using GitHub Desktop:
  - [Getting started with GitHub Desktop](https://docs.github.com/en/desktop/installing-and-configuring-github-desktop/getting-started-with-github-desktop) will guide you through setting up Desktop.
  - Once Desktop is set up, you can use it to [fork the repo](https://docs.github.com/en/desktop/contributing-and-collaborating-using-github-desktop/cloning-and-forking-repositories-from-github-desktop)!

- Using the command line:
  - [Fork the repo](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo#fork-an-example-repository) so that you can make your changes without affecting the original project until you're ready to merge them.

2. Install or make sure **Golang** is updated.

3. Create a working branch and start with your changes!

##### Recent Updates and Dependencies

Be aware of these recent changes when contributing:

- **HTTP Client Standardization**: All HTTP clients now use `httputil.DefaultClient` with custom User-Agent headers (`langchaingo/{version}`)
- **HuggingFace Environment Variables**: Supports multiple token sources in priority order: `HF_TOKEN`, `HUGGINGFACEHUB_API_TOKEN`, token file from `HF_TOKEN_PATH`, or default `~/.cache/huggingface/token`
- **OpenAI Functions Agent**: Updated to handle OpenAI's new tool calling API while maintaining backward compatibility
- **Chroma Vector Store**: Updated to use `github.com/amikos-tech/chroma-go` v0.1.4+
- **Testcontainers Migration**: New testcontainers API using `Run()` instead of deprecated `RunContainer()` where supported
- **HTTPRR Files**: No longer compressed - commit `.httprr` files directly to the repository

##### Project Structure and Conventions

When making changes, follow these architectural conventions:

- **HTTP Clients**: Use `httputil.DefaultClient` instead of `http.DefaultClient` for all HTTP operations to ensure proper User-Agent headers
- **Interface-based Design**: Core functionality is defined through interfaces (Model, Chain, Memory, etc.)
- **Provider Isolation**: Each LLM/embedding provider has its own package with internal client implementation
- **Options Pattern**: Use functional options for configuration (see existing examples)
- **Context Propagation**: All operations should accept `context.Context` for cancellation and deadlines
- **Error Handling**: Use standardized error types and mapping (see `llms.Error` and provider error mappers)

##### Adding a New LLM Provider

When adding a new LLM provider:

1. Create a new package under `/llms/your-provider`
2. Implement the `llms.Model` interface
3. Create an internal client package for HTTP interactions
4. Use `httputil.DefaultClient` for HTTP requests
5. Add compliance tests: `compliance.NewSuite("yourprovider", model).Run(t)`
6. Add tests with httprr recordings for HTTP calls
7. Follow the existing provider patterns for options and error handling

##### Adding a New Vector Store

When adding a new vector store:

1. Create a new package under `/vectorstores/your-store`
2. Implement the vector store interface
3. Use testcontainers for integration tests where possible
4. Follow existing patterns for distance strategies and metadata filtering

#### Running Tests

Before submitting your changes, make sure all tests pass:

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./chains

# Run a specific test
go test -run TestLLMChain ./chains

# Run tests with race detection
make test-race

# Run tests with coverage
make test-cover

# Test separation scripts
./scripts/run_unit_tests.sh      # Run only unit tests (no external dependencies)
./scripts/run_all_tests.sh       # Run complete test suite
./scripts/run_integration_tests.sh # Run only integration tests (requires Docker)

# Record HTTP interactions for tests (when adding new tests)
go test -httprecord=. -v ./path/to/package
```

Also ensure your code passes linting:

```bash
# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Run experimental linter configuration
make lint-exp

# Run all linters including experimental
make lint-all

# Clean lint cache
make clean-lint-cache

# Development tools
make build-examples         # Build all examples to verify they compile  
make docs                  # Generate documentation
make run-pkgsite          # Run local documentation server
make install-git-hooks    # Install git hooks (sets up pre-push hook)
make pre-push             # Run lint and fast tests (suitable for git pre-push hook)
```

##### Additional Development Tools

The project includes several development tools in `/internal/devtools`:

```bash
# Custom linting tools
make lint-devtools         # Run custom architectural lints
make lint-devtools-fix     # Run custom lints with auto-fix
make lint-architecture     # Run architectural validation
make lint-prepush          # Run pre-push lints
make lint-prepush-fix      # Run pre-push lints with auto-fix

# HTTPRR management
go run ./internal/devtools/rrtool list-packages  # List packages using httprr
make test-record           # Re-record all HTTP interactions

# Test pattern validation
make lint-testing          # Check for incorrect httprr test patterns
make lint-testing-fix      # Attempt to fix httprr test patterns automatically
```

#### Testing with httprr

This project uses a custom HTTP record/replay system (httprr) for testing HTTP interactions with external APIs. This allows tests to run deterministically without requiring actual API credentials or making real API calls.

##### How httprr works

- **Recording mode**: When tests run with real API credentials, httprr records all HTTP requests and responses to `.httprr` files in the `testdata` directory.
- **Replay mode**: When tests run without credentials, httprr replays the recorded HTTP interactions from the `.httprr` files.
- **Automatic mode switching**: Tests automatically skip if no credentials and no recording are available, with a helpful message.

##### Writing tests with httprr

When writing tests that make HTTP calls to external APIs, follow this pattern:

```go
func TestMyFeature(t *testing.T) {
    t.Parallel()
    ctx := context.Background()
    
    // Skip if no credentials and no recording
    httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")
    
    // Set up httprr (automatically cleaned up via t.Cleanup)
    // Use httputil.DefaultTransport for User-Agent headers, or http.DefaultTransport for simpler cases
    rr := httprr.OpenForTest(t, httputil.DefaultTransport)
    
    var opts []openai.Option
    opts = append(opts, openai.WithHTTPClient(rr.Client()))
    
    // Use test token when replaying
    if !rr.Recording() {
        opts = append(opts, openai.WithToken("test-api-key"))
    }
    // When recording, the client will use the real API key from environment
    
    client, err := openai.New(opts...)
    require.NoError(t, err)
    
    // Run your test
    result, err := client.Call(ctx, "test input")
    require.NoError(t, err)
    // ... assertions ...
}
```

This pattern ensures:
- **When recording**: Uses real API key from environment to capture valid responses
- **When replaying**: Uses "test-api-key" to satisfy client validation (httprr intercepts before actual API calls)

For other providers, use their specific options:

```go
// HuggingFace example (supports multiple environment variables)
func TestHuggingFace(t *testing.T) {
    // HuggingFace supports both HF_TOKEN and HUGGINGFACEHUB_API_TOKEN
    if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
        httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
    }
    
    rr := httprr.OpenForTest(t, httputil.DefaultTransport)
    
    apiKey := "test-api-key"
    if rr.Recording() {
        if key := os.Getenv("HF_TOKEN"); key != "" {
            apiKey = key
        } else if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" {
            apiKey = key
        }
    }
    
    llm, err := huggingface.New(
        huggingface.WithHTTPClient(rr.Client()),
        huggingface.WithToken(apiKey),
    )
    // ...
}

// Perplexity example  
var opts []perplexity.Option
opts = append(opts, perplexity.WithHTTPClient(rr.Client()))
if !rr.Recording() {
    opts = append(opts, perplexity.WithAPIKey("test-api-key"))
}
tool, err := perplexity.New(opts...)

// SerpAPI example with request scrubbing
rr.ScrubReq(func(req *http.Request) error {
    if req.URL != nil {
        q := req.URL.Query()
        q.Set("api_key", "test-api-key")
        req.URL.RawQuery = q.Encode()
    }
    return nil
})
```

For tests that need to create clients multiple times, consider using a helper function:

```go
func newOpenAILLM(t *testing.T) *openai.LLM {
    t.Helper()
    httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")
    
    rr := httprr.OpenForTest(t, httputil.DefaultTransport)
    
    // Only run tests in parallel when not recording (to avoid rate limits)
    if !rr.Recording() {
        t.Parallel()
    }
    
    var opts []openai.Option
    opts = append(opts, openai.WithHTTPClient(rr.Client()))
    
    if !rr.Recording() {
        opts = append(opts, openai.WithToken("test-api-key"))
    }
    // When recording, openai.New() will read OPENAI_API_KEY from environment
    
    llm, err := openai.New(opts...)
    require.NoError(t, err)
    return llm
}
```

##### Recording new tests

To record HTTP interactions for new tests:

1. Set the required environment variables (e.g., `OPENAI_API_KEY`)
2. Run the test with recording enabled:
   ```bash
   go test -v -httprecord=. ./path/to/package
   
   # To avoid rate limits, you can control parallelism:
   go test -v -httprecord=. -p 1 -parallel=1 ./path/to/package
   
   # Or use the Makefile target to record all packages
   make test-record
   ```
3. The test will create `.httprr` files in the `testdata` directory
4. Commit these recording files with your PR
5. For tests that require API key scrubbing, add request scrubbing functions

##### Important notes about httprr

- **Transport choice**: Use `httputil.DefaultTransport` for User-Agent headers, or `http.DefaultTransport` for simpler cases
- **Check rr.Recording()**: Use this to conditionally add test tokens only when replaying
- **httprr handles cleanup**: OpenForTest automatically registers cleanup with t.Cleanup()
- **Real keys for recording**: When recording, let the client use the real API key from environment
- **Test tokens for replay**: When replaying, use "test-api-key" to satisfy client validation
- **Parallel testing**: Only run `t.Parallel()` when not recording to avoid hitting API rate limits
- **Multiple credential sources**: For HuggingFace, check both `HF_TOKEN` and `HUGGINGFACEHUB_API_TOKEN`
- **Request scrubbing**: Use `rr.ScrubReq()` for APIs that need URL parameter scrubbing (like SerpAPI)
- **Recordings are deterministic**: The same inputs should produce the same outputs
- **Sensitive data is scrubbed**: httprr automatically removes authorization headers and other sensitive data from recordings
- **Commit recording files**: Always commit the `.httprr` files so tests can run in CI without credentials
- **Delete invalid recordings**: If a test fails due to an invalid recording (e.g., 401 error), delete the recording file and re-record with valid credentials

##### Debugging httprr issues

- Use `-httprecord-debug` flag for detailed recording information
- Use `-httpdebug` flag to see actual HTTP traffic
- Check if recordings exist: `ls testdata/*.httprr`
- Verify recording contents: `head testdata/TestName.httprr`
- Use test separation scripts to isolate unit vs integration test issues:
  ```bash
  ./scripts/run_unit_tests.sh      # Fast tests without external dependencies
  ./scripts/run_integration_tests.sh # Tests requiring Docker/external services
  ```

##### Automated httprr pattern validation

The project includes a custom linter to detect incorrect httprr usage patterns:

```bash
# Check for incorrect patterns
make lint-testing

# See specific issues found
go run ./internal/devtools/lint -testing -v
```

The linter detects:
- **Hardcoded test tokens**: `WithToken("test-api-key")` called unconditionally (should be conditional on `!rr.Recording()`)  
- **Incorrect parallel execution**: `t.Parallel()` called before httprr setup (should be conditional on `!rr.Recording()`)

These issues cause authentication errors during recording and race conditions during testing.

#### Commit your update

Commit the changes once you are happy with them. Don't forget to self-review to speed up the review process:zap:.

#### Pull Request

When you're finished with the changes, create a pull request, also known as a PR.
- Name your Pull Request title clearly, concisely, and prefixed with the name of primarily affected package you changed according to [Go Contribute Guideline](https://go.dev/doc/contribute#commit_messages). (such as `memory: add interfaces` or `util: add helpers`)
- Run all linters and ensure tests pass: `make lint && make test`
- If you added new HTTP-based functionality, include httprr recordings
- **We strive to conceptually align with the Python and TypeScript versions of Langchain. Please link/reference the associated concepts in those codebases when introducing a new concept.**
- Fill the "Ready for review" template so that we can review your PR. This template helps reviewers understand your changes as well as the purpose of your pull request.
- Don't forget to [link PR to issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue) if you are solving one.
- Enable the checkbox to [allow maintainer edits](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/allowing-changes-to-a-pull-request-branch-created-from-a-fork) so the branch can be updated for a merge.
Once you submit your PR, a team member will review your proposal. We may ask questions or request additional information.
- We may ask for changes to be made before a PR can be merged, either using [suggested changes](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/incorporating-feedback-in-your-pull-request) or pull request comments. You can apply suggested changes directly through the UI. You can make any other changes in your fork, then commit them to your branch.
- As you update your PR and apply changes, mark each conversation as [resolved](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/commenting-on-a-pull-request#resolving-conversations).
- If you run into any merge issues, checkout this [git tutorial](https://github.com/skills/resolve-merge-conflicts) to help you resolve merge conflicts and other issues.

#### Your PR is merged!

Congratulations :tada::tada: The langchaingo team thanks you :sparkles:.

Once your PR is merged, your contributions will be publicly visible on the repository contributors list.

Now that you are part of the community!
