package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// WeatherReading represents a single dummy weather data record.
type WeatherReading struct {
	City        string    `json:"city"`
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"` // Celsius
	Humidity    int       `json:"humidity"`    // Percentage
	Condition   string    `json:"condition"`
}

// DataResponse holds the array of weather readings.
type DataResponse struct {
	Readings []WeatherReading `json:"readings"`
	Message  string           `json:"message,omitempty"` // Added for error messages
}

// DataResponse holds the array of weather readings.
type DataResponse struct {
	Readings []WeatherReading `json:"readings"`
	Message  string           `json:"message,omitempty"` // Added for error messages
}

// Global random source for generating values and status codes.
var r *rand.Rand

func init() {
	// Initialize a new Rand source with the current time for better randomness.
	s := rand.NewSource(time.Now().UnixNano())
	r = rand.New(s)
}

// generateDummyWeatherReadings generates a slice of dummy WeatherReading objects.
func generateDummyWeatherReadings(count int) []WeatherReading {
	readings := make([]WeatherReading, count)
	cities := []string{"New York", "London", "Paris", "Tokyo", "Sydney", "Lagos", "Dubai", "Rio"}
	conditions := []string{"Sunny", "Partly Cloudy", "Cloudy", "Rainy", "Stormy", "Foggy", "Snowy"}

	for i := 0; i < count; i++ {
		readings[i] = WeatherReading{
			City:        cities[r.Intn(len(cities))],
			Timestamp:   time.Now().Add(time.Duration(r.Intn(24)-12) * time.Hour), // Simulate readings +/- 12 hours
			Temperature: float64(r.Intn(35)+5) + r.Float64(),                      // 5.0 to 40.0 Celsius
			Humidity:    r.Intn(80) + 20,                                          // 20% to 99%
			Condition:   conditions[r.Intn(len(conditions))],
		}
	}
	return readings
}

// getResponseStatusCode randomly selects a 2xx, 4xx, or 5xx status code.
func getResponseStatusCode() int {
	statusCodes2xx := []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent}
	statusCodes4xx := []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusForbidden, http.StatusMethodNotAllowed}
	statusCodes5xx := []int{http.StatusInternalServerError, http.StatusNotImplemented, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout}

	// Randomly decide the type of response: 2xx, 4xx, or 5xx
	// Adjust weights if needed, e.g., 70% 2xx, 15% 4xx, 15% 5xx
	randomNumber := r.Intn(100) // 0-99
	if randomNumber < 70 {      // 70% chance for 2xx
		return statusCodes2xx[r.Intn(len(statusCodes2xx))]
	} else if randomNumber < 85 { // 15% chance for 4xx (80-89)
		return statusCodes4xx[r.Intn(len(statusCodes4xx))]
	} else { // 15% chance for 5xx (90-99)
		return statusCodes5xx[r.Intn(len(statusCodes5xx))]
	}
}

// weatherHandler handles requests to the /weather endpoint.
// It now takes a Sleeper interface for dependency injection.
func weatherHandler(s Sleeper, w http.ResponseWriter, req *http.Request) {
	// Set Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Get response size from query parameter, default to 10 if not provided or invalid.
	sizeStr := req.URL.Query().Get("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 10 || size > 100 {
		log.Printf("Invalid or missing 'size' parameter, defaulting to 10. Received: %s", sizeStr)
		size = 10 // Default size
	}

	// Introduce a random delay between 0 and 5 seconds using the injected Sleeper.
	delay := time.Duration(r.Intn(5001)) * time.Millisecond // 0 to 5000 milliseconds
	log.Printf("Introducing a delay of %v for this request.", delay)
	s.Sleep(delay) // Use the injected sleeper

	// Get a random status code
	statusCode := getResponseStatusCode()
	w.WriteHeader(statusCode)

	var responseData DataResponse

	// Depending on the status code, provide appropriate response body
	if statusCode >= 200 && statusCode < 300 {
		readings := generateDummyWeatherReadings(size)
		responseData = DataResponse{
			Readings: readings,
			Message:  fmt.Sprintf("Successfully retrieved %d weather readings.", len(readings)),
		}
		log.Printf("Responding with %d status code and %d weather readings.", statusCode, len(readings))
	} else {
		// For 4xx and 5xx errors, provide a generic error message.
		errorMessage := fmt.Sprintf("An error occurred with status code %d. This is a dummy error for testing.", statusCode)
		responseData = DataResponse{
			Message: errorMessage,
		}
		log.Printf("Responding with %d status code and error message: %s", statusCode, errorMessage)
	}

	// Encode and send the JSON response
	json.NewEncoder(w).Encode(responseData)
}

func health(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Healthy")) }

func main() {
	// Create an instance of RealSleeper for the main application.
	sleeper := &DefaultSleeper{}

	// Define the handler for the /weather endpoint, injecting the realSleeper.
	http.HandleFunc("/weather", func(w http.ResponseWriter, req *http.Request) {
		weatherHandler(sleeper, w, req)
	})
	http.HandleFunc("/health", health)

	// Start the HTTP server
	port := ":8080"
	log.Printf("Starting Go REST API server on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
