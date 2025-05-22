// Package examples demonstrates how to use httprr for offline testing.
package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/internal/httprr"
)

// WeatherAPI is a simple client for a weather API.
type WeatherAPI struct {
	client *http.Client
}

// WeatherResponse represents the response from the weather API.
type WeatherResponse struct {
	Temperature float64 `json:"temperature"`
	Location    string  `json:"location"`
	Unit        string  `json:"unit"`
}

// NewWeatherAPI creates a new WeatherAPI client.
func NewWeatherAPI(client *http.Client) *WeatherAPI {
	if client == nil {
		client = http.DefaultClient
	}
	return &WeatherAPI{client: client}
}

// GetWeather fetches the current weather for a location.
func (w *WeatherAPI) GetWeather(ctx context.Context, location string, unit string) (*WeatherResponse, error) {
	// This would normally be a real API endpoint
	url := fmt.Sprintf("https://api.example.com/weather?location=%s&unit=%s", location, unit)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, err
	}
	
	return &weatherResp, nil
}

// createMockRecording creates a mock recording file for testing
func createMockRecording(t *testing.T, recordingsDir string) {
	// Create a simplified mock response that works with our parser
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`{"temperature":65.5,"location":"Munich","unit":"fahrenheit"}`)),
	}
	
	// Create a simple request
	req, _ := http.NewRequest("GET", "https://api.example.com/weather?location=Munich&unit=fahrenheit", nil)
	
	// Create a recorder to save the request/response
	recorder := httprr.NewRecorder(nil)
	recorder.Dir = recordingsDir
	recorder.Mode = httprr.ModeRecord
	
	// Save the request/response pair
	// We're creating a record manually rather than making a real request
	record := &httprr.Record{
		Request:  req,
		Response: resp,
	}
	
	// Save the record to disk
	if err := recorder.SaveRecord(record); err != nil {
		t.Fatalf("Failed to save mock recording: %v", err)
	}
}

// TestWeatherAPIOffline demonstrates offline testing with httprr.
func TestWeatherAPIOffline(t *testing.T) {
	// Create a test helper for mocking HTTP requests
	helper := httprr.NewTestHelper(t)
	defer helper.Cleanup()
	
	// Create a manual recorder response
	responseJSON := `{"temperature":65.5,"location":"Munich","unit":"fahrenheit"}`
	
	// Create custom roundtripper that returns our predefined response
	helper.Client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Check that the request is for the weather API
		if strings.Contains(req.URL.String(), "api.example.com/weather") {
			return &http.Response{
				StatusCode: 200,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(responseJSON)),
				Header:     make(http.Header),
			}, nil
		}
		return nil, fmt.Errorf("unexpected request: %s", req.URL)
	})
	
	// Create the weather API with our custom client
	api := NewWeatherAPI(helper.Client)
	
	// Call the API - this should use our mocked response
	weather, err := api.GetWeather(context.Background(), "Munich", "fahrenheit")
	if err != nil {
		t.Fatalf("Failed to get weather: %v", err)
	}
	
	// Verify the response matches our expectations
	if weather.Location != "Munich" {
		t.Errorf("Expected location to be Munich, got %s", weather.Location)
	}
	if weather.Temperature != 65.5 {
		t.Errorf("Expected temperature to be 65.5, got %f", weather.Temperature)
	}
	if weather.Unit != "fahrenheit" {
		t.Errorf("Expected unit to be fahrenheit, got %s", weather.Unit)
	}
}

// roundTripperFunc adapter for simple response mocking
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestWeatherAPIWithEnvironmentControl demonstrates using the environment to control recording mode.
func TestWeatherAPIWithEnvironmentControl(t *testing.T) {
	// Create a recordings directory for this test
	recordingsDir := "testdata/weather-recordings"
	
	// Create a test helper that respects environment variables
	helper := httprr.NewAutoHelper(t, recordingsDir)
	defer helper.Cleanup()
	
	// Create the weather API with our configured client
	api := NewWeatherAPI(helper.Client)
	
	// First run (in record mode) will create recordings
	// Second+ runs (in replay mode) will use those recordings
	
	// If no recordings exist yet, create a mock one to demonstrate replay mode
	if helper.Recorder.Mode == httprr.ModeRecord {
		// This would only happen on first run or when HTTPRR_MODE=record
		t.Logf("Running in RECORD mode")
		createMockRecording(t, recordingsDir)
	} else {
		t.Logf("Running in REPLAY mode")
	}
	
	// Call the API - this will use the real API if in record mode,
	// or the recorded response if in replay mode
	weather, err := api.GetWeather(context.Background(), "Munich", "fahrenheit")
	if err != nil {
		t.Fatalf("Failed to get weather: %v", err)
	}
	
	// Verify the response matches our expectations
	if weather.Location != "Munich" {
		t.Errorf("Expected location to be Munich, got %s", weather.Location)
	}
	if weather.Temperature != 65.5 {
		t.Errorf("Expected temperature to be 65.5, got %f", weather.Temperature)
	}
	if weather.Unit != "fahrenheit" {
		t.Errorf("Expected unit to be fahrenheit, got %s", weather.Unit)
	}
	
	// Print info about recorded requests for debugging
	h := helper.Recorder.Records()
	t.Logf("Recorded %d HTTP interactions", len(h))
}
