package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid http URL",
			url:         "http://example.com",
			expectError: false,
		},
		{
			name:        "valid https URL",
			url:         "https://example.com",
			expectError: false,
		},
		{
			name:        "valid URL with path",
			url:         "https://example.com/path/to/resource",
			expectError: false,
		},
		{
			name:        "valid URL with query parameters",
			url:         "https://example.com?foo=bar&baz=qux",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
			errorMsg:    "URL is required",
		},
		{
			name:        "whitespace only URL",
			url:         "   ",
			expectError: true,
			errorMsg:    "URL is required",
		},
		{
			name:        "URL without scheme",
			url:         "example.com",
			expectError: true,
			errorMsg:    "URL must include a scheme (http or https)",
		},
		{
			name:        "URL with invalid scheme",
			url:         "ftp://example.com",
			expectError: true,
			errorMsg:    "URL scheme must be http or https",
		},
		{
			name:        "URL without host",
			url:         "http://",
			expectError: true,
			errorMsg:    "URL must include a host",
		},
		{
			name:        "malformed URL",
			url:         "http://[invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateURL(tt.url)

			if tt.expectError {
				if result == "" {
					t.Errorf("expected error but got none")
				}
				if tt.errorMsg != "" && result != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, result)
				}
			} else {
				if result != "" {
					t.Errorf("expected no error but got: %s", result)
				}
			}
		})
	}
}

func TestIsBlocked(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "403 Forbidden",
			statusCode: 403,
			expected:   true,
		},
		{
			name:       "429 Too Many Requests",
			statusCode: 429,
			expected:   true,
		},
		{
			name:       "200 OK",
			statusCode: 200,
			expected:   false,
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			expected:   false,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBlocked(tt.statusCode)
			if result != tt.expected {
				t.Errorf("isBlocked(%d) = %v, expected %v", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestTestURL(t *testing.T) {
	// Test successful request
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		}))
		defer server.Close()

		response := testURL(server.URL)

		if !response.Success {
			t.Errorf("expected success, got failure: %s", response.Error)
		}
		if response.StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", response.StatusCode)
		}
		if response.ResponseTime < 0 {
			t.Errorf("expected non-negative response time, got %d", response.ResponseTime)
		}
		if response.BodyPreview != "Hello, World!" {
			t.Errorf("expected body preview 'Hello, World!', got %q", response.BodyPreview)
		}
		if response.Truncated {
			t.Errorf("expected truncated to be false, got true")
		}
		if response.Blocked {
			t.Errorf("expected blocked to be false, got true")
		}
	})

	// Test blocked request (403)
	t.Run("blocked request 403", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		response := testURL(server.URL)

		if !response.Success {
			t.Errorf("expected success, got failure: %s", response.Error)
		}
		if response.StatusCode != 403 {
			t.Errorf("expected status code 403, got %d", response.StatusCode)
		}
		if !response.Blocked {
			t.Errorf("expected blocked to be true, got false")
		}
	})

	// Test blocked request (429)
	t.Run("blocked request 429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		response := testURL(server.URL)

		if !response.Success {
			t.Errorf("expected success, got failure: %s", response.Error)
		}
		if response.StatusCode != 429 {
			t.Errorf("expected status code 429, got %d", response.StatusCode)
		}
		if !response.Blocked {
			t.Errorf("expected blocked to be true, got false")
		}
	})

	// Test body truncation
	t.Run("body truncation", func(t *testing.T) {
		longBody := ""
		for i := 0; i < 2000; i++ {
			longBody += "a"
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(longBody))
		}))
		defer server.Close()

		response := testURL(server.URL)

		if !response.Success {
			t.Errorf("expected success, got failure: %s", response.Error)
		}
		if len(response.BodyPreview) != 1000 {
			t.Errorf("expected body preview length 1000, got %d", len(response.BodyPreview))
		}
		if !response.Truncated {
			t.Errorf("expected truncated to be true, got false")
		}
	})

	// Test User-Agent header
	t.Run("user agent header", func(t *testing.T) {
		var userAgent string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userAgent = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		testURL(server.URL)

		if userAgent == "" {
			t.Errorf("expected User-Agent header to be set")
		}
		if userAgent != "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" {
			t.Errorf("expected specific User-Agent, got %q", userAgent)
		}
	})

	// Test redirect tracking
	t.Run("redirect tracking", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/redirect" {
				http.Redirect(w, r, "/final", http.StatusMovedPermanently)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Final destination"))
			}
		}))
		defer server.Close()

		response := testURL(server.URL + "/redirect")

		if !response.Success {
			t.Errorf("expected success, got failure: %s", response.Error)
		}
		if response.FinalURL != server.URL+"/final" {
			t.Errorf("expected final URL %q, got %q", server.URL+"/final", response.FinalURL)
		}
	})

	// Test invalid URL
	t.Run("invalid URL", func(t *testing.T) {
		response := testURL("http://invalid.example.test.invalid.local")

		if response.Success {
			t.Errorf("expected failure for invalid URL")
		}
		if response.Error == "" {
			t.Errorf("expected error message for invalid URL")
		}
	})
}

func TestTestURLHandler(t *testing.T) {
	// Test successful API request
	t.Run("successful API request", func(t *testing.T) {
		// Create a test server that will be the target URL
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Test response"))
		}))
		defer targetServer.Close()

		// Create a request to the API endpoint
		reqBody := `{"url":"` + targetServer.URL + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Record the response
		w := httptest.NewRecorder()
		testURLHandler(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("expected status code 200, got %d", w.Code)
		}

		var response TestResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success, got failure: %s", response.Error)
		}
		if response.StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", response.StatusCode)
		}
	})

	// Test validation error (empty URL)
	t.Run("validation error - empty URL", func(t *testing.T) {
		reqBody := `{"url":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		testURLHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code 400, got %d", w.Code)
		}

		var errResponse map[string]string
		if err := json.NewDecoder(w.Body).Decode(&errResponse); err != nil {
			t.Errorf("failed to decode error response: %v", err)
		}

		if errResponse["error"] != "URL is required" {
			t.Errorf("expected error message 'URL is required', got %q", errResponse["error"])
		}
	})

	// Test validation error (invalid URL format)
	t.Run("validation error - invalid URL format", func(t *testing.T) {
		reqBody := `{"url":"not-a-url"}`
		req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		testURLHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code 400, got %d", w.Code)
		}

		var errResponse map[string]string
		if err := json.NewDecoder(w.Body).Decode(&errResponse); err != nil {
			t.Errorf("failed to decode error response: %v", err)
		}

		if errResponse["error"] == "" {
			t.Errorf("expected error message, got empty string")
		}
	})

	// Test invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		reqBody := `{invalid json}`
		req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		testURLHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code 400, got %d", w.Code)
		}

		var errResponse map[string]string
		if err := json.NewDecoder(w.Body).Decode(&errResponse); err != nil {
			t.Errorf("failed to decode error response: %v", err)
		}

		if errResponse["error"] != "Invalid JSON" {
			t.Errorf("expected error message 'Invalid JSON', got %q", errResponse["error"])
		}
	})

	// Test method not allowed
	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()
		testURLHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status code 405, got %d", w.Code)
		}
	})
}

func TestHealthHandler(t *testing.T) {
	// Test successful health check
	t.Run("health check returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		healthHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code 200, got %d", w.Code)
		}

		if w.Body.String() != "OK" {
			t.Errorf("expected response body 'OK', got %q", w.Body.String())
		}
	})

	// Test method not allowed
	t.Run("health check method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()
		healthHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status code 405, got %d", w.Code)
		}
	})
}
