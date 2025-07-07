package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var sleeper = &NoOpSleeper{}

// TestWeatherHandlerSuccess tests the /weather endpoint for successful responses (2xx).
func TestWeatherHandlerSuccess(t *testing.T) {
	// Test with default size (10)
	req := httptest.NewRequest("GET", "/weather", nil)
	rr := httptest.NewRecorder()
	weatherHandler(sleeper, rr, req)

	// Check Content-Type header
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	// Parse the response body
	var responseData DataResponse
	err := json.NewDecoder(rr.Body).Decode(&responseData)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Check if it's a 2xx status code (since it's random, we can't predict exact, but expect success path)
	// For a more deterministic test of 2xx, 4xx, 5xx, you'd need to mock the random number generator.
	// Here, we assume a successful path for this test case.
	if rr.Code < 200 || rr.Code >= 300 {
		t.Logf("Warning: Handler returned non-2xx status code %d in success test. This is due to randomness.", rr.Code)
		// If it's an error, the readings list will be nil, so we can't check its length.
		// We'll proceed to check message presence.
		if responseData.Message == "" {
			t.Errorf("Expected a message in error response, but got empty.")
		}
		return // Exit if it wasn't a 2xx, as the rest of the checks are for 2xx.
	}

	// Check if readings are present and their count is default (10)
	if responseData.Readings == nil {
		t.Errorf("Expected readings in successful response, but got nil.")
	}
	if len(responseData.Readings) != 10 {
		t.Errorf("Handler returned unexpected number of readings: got %d want %d (default size)", len(responseData.Readings), 10)
	}
	if responseData.Message == "" {
		t.Errorf("Expected a message in successful response, but got empty.")
	}

	// Test with a specific valid size (e.g., 50)
	req = httptest.NewRequest("GET", "/weather?size=50", nil)
	rr = httptest.NewRecorder()
	weatherHandler(sleeper, rr, req)

	err = json.NewDecoder(rr.Body).Decode(&responseData)
	if err != nil {
		t.Fatalf("Could not decode response for size=50: %v", err)
	}

	if rr.Code < 200 || rr.Code >= 300 {
		t.Logf("Warning: Handler returned non-2xx status code %d for size=50 test. This is due to randomness.", rr.Code)
		return
	}

	if len(responseData.Readings) != 50 {
		t.Errorf("Handler returned unexpected number of readings for size=50: got %d want %d", len(responseData.Readings), 50)
	}

	// Test with max size (100)
	req = httptest.NewRequest("GET", "/weather?size=100", nil)
	rr = httptest.NewRecorder()
	weatherHandler(sleeper, rr, req)

	err = json.NewDecoder(rr.Body).Decode(&responseData)
	if err != nil {
		t.Fatalf("Could not decode response for size=100: %v", err)
	}

	if rr.Code < 200 || rr.Code >= 300 {
		t.Logf("Warning: Handler returned non-2xx status code %d for size=100 test. This is due to randomness.", rr.Code)
		return
	}

	if len(responseData.Readings) != 100 {
		t.Errorf("Handler returned unexpected number of readings for size=100: got %d want %d", len(responseData.Readings), 100)
	}
}

// TestWeatherHandlerInvalidSize tests the /weather endpoint with invalid size parameters.
func TestWeatherHandlerInvalidSize(t *testing.T) {
	testCases := []struct {
		name      string
		sizeParam string
	}{
		{"SizeTooSmall", "5"},
		{"SizeTooLarge", "150"},
		{"NonNumericSize", "abc"},
		{"EmptySize", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/weather?size="+tc.sizeParam, nil)
			rr := httptest.NewRecorder()
			weatherHandler(sleeper, rr, req)

			// Parse the response body
			var responseData DataResponse
			err := json.NewDecoder(rr.Body).Decode(&responseData)
			if err != nil {
				t.Fatalf("Could not decode response: %v", err)
			}

			// Regardless of the random status code, the size should default to 10
			// if the parameter is invalid.
			if rr.Code >= 200 && rr.Code < 300 {
				if responseData.Readings == nil {
					t.Errorf("Expected readings in successful response for invalid size, but got nil.")
				}
				if len(responseData.Readings) != 10 {
					t.Errorf("Handler returned unexpected number of readings for invalid size '%s': got %d want %d (default)", tc.sizeParam, len(responseData.Readings), 10)
				}
			}
			// For error responses, we just check if a message is present.
			if responseData.Message == "" {
				t.Errorf("Expected a message in response for invalid size, but got empty.")
			}
		})
	}
}

// TestWeatherHandlerErrorResponseStructure tests that error responses have a message and no readings.
func TestWeatherHandlerErrorResponseStructure(t *testing.T) {
	// This test relies on the randomness to eventually hit a 4xx or 5xx.
	// For a more robust test, you'd mock `getResponseStatusCode`.
	// We'll make multiple attempts to increase the chance of hitting an error.
	maxAttempts := 10
	errorHit := false

	for i := 0; i < maxAttempts; i++ {
		req := httptest.NewRequest("GET", "/weather", nil)
		rr := httptest.NewRecorder()
		weatherHandler(sleeper, rr, req)

		if rr.Code >= 400 { // Check for 4xx or 5xx status codes
			errorHit = true
			var responseData DataResponse
			err := json.NewDecoder(rr.Body).Decode(&responseData)
			if err != nil {
				t.Fatalf("Could not decode error response: %v", err)
			}

			if len(responseData.Readings) > 0 {
				t.Errorf("Error response should not contain readings, but found %d.", len(responseData.Readings))
			}
			if responseData.Message == "" {
				t.Errorf("Error response should contain a message, but it was empty.")
			}
			t.Logf("Successfully tested error response structure for status code %d.", rr.Code)
			break // Exit loop once an error response is hit
		}
	}

	if !errorHit {
		t.Log("Warning: Did not hit an error status code after multiple attempts. Consider increasing maxAttempts or mocking randomness for deterministic error testing.")
	}
}

// TestHealthEndpoint tests the /health endpoint.
func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Health endpoint returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	expected := "Healthy"
	if rr.Body.String() != expected {
		t.Errorf("Health endpoint returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
