# httprr - HTTP Request/Response Recorder

`httprr` is a package for recording and replaying HTTP traffic in tests. The name is inspired by "http recorder/replayer" similar to the way Russ Cox would name packages.

## Overview

This package provides tools for:

1. Recording HTTP requests and responses during tests
2. Saving these recordings to disk for inspection or future replay
3. Replaying recorded HTTP interactions for offline testing
4. Asserting on HTTP interactions in tests
5. Debugging HTTP traffic by dumping detailed logs

## Key Features

- Records complete HTTP requests and responses, including headers and bodies
- Preserves the original request/response for inspection
- Supports offline testing by replaying previously recorded responses
- Provides detailed dumps of HTTP interactions for debugging
- Simple API for test assertions
- Supports saving recordings to disk
- Minimal overhead when recording HTTP traffic

## Usage

### Basic Usage in a Test

```go
func TestMyFunction(t *testing.T) {
    // Create the test helper
    httpHelper := httprr.NewTestHelper(t)
    defer httpHelper.Cleanup()

    // Use the client with your code under test
    myService := service.New(httpHelper.Client)
    
    // Run your test
    result := myService.DoSomething()
    
    // Assert on the result
    require.Equal(t, "expected", result)
    
    // Assert on the HTTP interactions
    httpHelper.AssertRequestCount(1)
    httpHelper.AssertURLCalled("api.example.com")
    
    // Dump recordings for debugging (only in verbose mode)
    if testing.Verbose() {
        httpHelper.DumpRecordings()
    }
}
```

### Replay Mode for Offline Testing

```go
func TestMyFunctionOffline(t *testing.T) {
    // Create a test helper that uses replay mode with a specific directory
    recordingsDir := filepath.Join("testdata", "my_test_recordings")
    httpHelper := httprr.NewReplayHelper(t, recordingsDir)
    defer httpHelper.Cleanup() // Won't delete permanent recordings

    // Use the client with your code under test
    myService := service.New(httpHelper.Client)
    
    // Run your test - this will use recorded responses instead of making real HTTP calls
    result := myService.DoSomething()
    
    // Assert on the result
    require.Equal(t, "expected", result)
}
```

### Auto-Mode (Record or Replay)

```go
func TestMyFunction(t *testing.T) {
    // Create a test helper that automatically chooses record or replay mode
    recordingsDir := filepath.Join("testdata", "my_test_recordings")
    httpHelper := httprr.NewAutoHelper(t, recordingsDir)
    
    // Use the client with your code under test
    myService := service.New(httpHelper.Client)
    
    // Run your test - will use recorded responses if available, otherwise make real HTTP calls
    result := myService.DoSomething()
    
    // Assert on the result
    require.Equal(t, "expected", result)
}
```

Auto mode determines whether to record or replay based on:
1. The `HTTPRR_MODE` environment variable (set to "record" or "replay")
2. Whether the recordings directory already exists (if it does, defaults to replay)

### Direct Usage of the Recorder

```go
func TestCustomRecorder(t *testing.T) {
    // Create a recorder with custom settings
    recorder := httprr.NewRecorder(http.DefaultTransport)
    recorder.Dir = "testdata/recordings"
    recorder.Mode = httprr.ModeRecord // or ModeReplay for offline testing
    
    // Create a client with the recorder
    client := &http.Client{Transport: recorder}
    
    // Make HTTP requests
    resp, err := client.Get("https://api.example.com")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Access the recordings
    records := recorder.Records()
    require.Len(t, records, 1)
    require.Equal(t, "GET", records[0].Request.Method)
}
```

### Creating a Replay Client Directly

```go
// Create a client that will replay responses from a directory
recordingsDir := "testdata/recordings"
client := httprr.ReplayClient(recordingsDir, http.DefaultTransport)

// Use the client - it will use recorded responses instead of making real HTTP calls
resp, err := client.Get("https://api.example.com")
```

## Running Tests Offline

To run tests without making real HTTP calls, you can:

1. Record mode: First run tests to record HTTP interactions
```bash
# Record mode (default)
go test ./...
```

2. Replay mode: Then run tests offline using recorded responses
```bash
# Replay mode
HTTPRR_MODE=replay go test ./...
```

## Migrating Tests to Use httprr

When migrating tests to use httprr, replace uses of `http.DefaultClient` or custom clients with the httprr client:

Before:
```go
func TestExample(t *testing.T) {
    service := myservice.New(http.DefaultClient)
    // ... rest of test
}
```

After:
```go
func TestExample(t *testing.T) {
    httpHelper := httprr.NewAutoHelper(t, filepath.Join("testdata", "example_recordings"))
    
    service := myservice.New(httpHelper.Client)
    // ... rest of test
    
    // Add assertions on HTTP interactions
    httpHelper.AssertURLCalled("api.example.com")
}
```

## Design Philosophy

The design of this package follows these principles:

1. **Simplicity**: Easy to use with minimal setup
2. **Transparency**: Record HTTP interactions without affecting functionality
3. **Testability**: Make assertions about HTTP interactions
4. **Debugging**: Provide rich information for debugging
5. **Offline Testing**: Enable tests to run without network access
6. **Minimal Dependencies**: Relies mostly on standard library

## Contributing

When adding features to this package, please maintain backward compatibility and add tests for any new functionality. 