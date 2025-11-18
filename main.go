package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Global variable to cache server IP
var (
	cachedServerIP string
	ipMutex        sync.RWMutex
)

// TestRequest represents a URL test request from the client
type TestRequest struct {
	URL string `json:"url"`
}

// TestResponse represents the result of a URL test
type TestResponse struct {
	Success      bool              `json:"success"`
	StatusCode   int               `json:"statusCode,omitempty"`
	ResponseTime int64             `json:"responseTime,omitempty"` // milliseconds
	FinalURL     string            `json:"finalUrl,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	BodyPreview  string            `json:"bodyPreview,omitempty"`
	Truncated    bool              `json:"truncated"`
	Error        string            `json:"error,omitempty"`
	Blocked      bool              `json:"blocked"`
	UserIP       string            `json:"userIP,omitempty"`
	ServerIP     string            `json:"serverIP,omitempty"`
}

// validateURL checks if a URL is valid
// Returns an error message if the URL is invalid, or empty string if valid
func validateURL(urlStr string) string {
	// Check if URL is empty or only whitespace
	if strings.TrimSpace(urlStr) == "" {
		return "URL is required"
	}

	// Parse the URL to check format validity
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "Invalid URL format: " + err.Error()
	}

	// Check if scheme is present (http or https)
	if parsedURL.Scheme == "" {
		return "URL must include a scheme (http or https)"
	}

	// Check if scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "URL scheme must be http or https"
	}

	// Check if host is present
	if parsedURL.Host == "" {
		return "URL must include a host"
	}

	return ""
}

func main() {
	// Fetch server IP on startup (in background to not block startup)
	go func() {
		ip := fetchServerIP()
		ipMutex.Lock()
		cachedServerIP = ip
		ipMutex.Unlock()
		log.Printf("Server IP: %s", ip)
	}()

	// Set up routes
	http.HandleFunc("/", serveStaticHandler)
	http.HandleFunc("/api/test", testURLHandler)
	http.HandleFunc("/health", healthHandler)

	// Get PORT from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Log startup message to stdout
	log.Printf("URL Tester starting on port %s", port)

	// Start HTTP server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// healthHandler handles GET /health requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// serveStaticHandler serves static files from the static directory
func serveStaticHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Serve files from static directory
	fs := http.FileServer(http.Dir("./static"))
	fs.ServeHTTP(w, r)
}

// testURLHandler handles POST /api/test requests
func testURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate URL
	if validationErr := validateURL(req.URL); validationErr != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": validationErr})
		return
	}

	// Test the URL
	response := testURL(req.URL)

	// Add user IP and server IP to response
	response.UserIP = getClientIP(r)
	response.ServerIP = getServerIP()

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// createHTTPClient creates a custom HTTP client with 30-second timeout
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects by returning nil
			return nil
		},
	}
}

// formatError formats an error message for display to the user
func formatError(err error) string {
	if err == nil {
		return ""
	}

	// Check for timeout error
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return "timeout: request exceeded 30 seconds"
	}

	// Check for SSL/TLS errors
	if _, ok := err.(*tls.CertificateVerificationError); ok {
		return "SSL/TLS error: certificate verification failed"
	}

	// Check for other SSL/TLS errors
	if _, ok := err.(tls.RecordHeaderError); ok {
		return "SSL/TLS error: invalid certificate or protocol error"
	}

	// Check for DNS errors
	if dnsErr, ok := err.(*net.DNSError); ok {
		if dnsErr.IsNotFound {
			return fmt.Sprintf("DNS error: host not found (%s)", dnsErr.Name)
		}
		return fmt.Sprintf("DNS error: %s", dnsErr.Err)
	}

	// Check for connection errors
	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Op == "dial" {
			if syscallErr, ok := opErr.Err.(interface{ Error() string }); ok {
				return fmt.Sprintf("connection error: %s", syscallErr.Error())
			}
			return "connection error: failed to connect to host"
		}
	}

	// Generic error message
	errStr := err.Error()
	// Truncate very long error messages
	if len(errStr) > 200 {
		errStr = errStr[:200] + "..."
	}
	return errStr
}

// testURL sends an HTTP request to the target URL and returns the result
func testURL(targetURL string) TestResponse {
	client := createHTTPClient()

	// Create request
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		errMsg := formatError(err)
		fmt.Fprintf(os.Stderr, "Error creating request for URL %s: %v\n", targetURL, err)
		return TestResponse{
			Success: false,
			Error:   errMsg,
			Blocked: false,
		}
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Record start time
	startTime := time.Now()

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		// Log error to stderr
		errMsg := formatError(err)
		fmt.Fprintf(os.Stderr, "Error testing URL %s: %v\n", targetURL, err)
		return TestResponse{
			Success: false,
			Error:   errMsg,
			Blocked: false,
		}
	}
	defer resp.Body.Close()

	// Calculate response time in milliseconds
	responseTime := time.Since(startTime).Milliseconds()

	// Extract headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Read response body (limited to 1000 characters)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body for %s: %v\n", targetURL, err)
		return TestResponse{
			Success:    true,
			StatusCode: resp.StatusCode,
			FinalURL:   resp.Request.URL.String(),
			Headers:    headers,
			Blocked:    isBlocked(resp.StatusCode),
		}
	}

	bodyStr := string(bodyBytes)
	truncated := false
	bodyPreview := bodyStr

	// Truncate body if necessary
	if len(bodyStr) > 1000 {
		bodyPreview = bodyStr[:1000]
		truncated = true
	}

	// Check if blocked
	blocked := isBlocked(resp.StatusCode)

	return TestResponse{
		Success:      true,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		FinalURL:     resp.Request.URL.String(),
		Headers:      headers,
		BodyPreview:  bodyPreview,
		Truncated:    truncated,
		Blocked:      blocked,
	}
}

// isBlocked checks if the response indicates the request was blocked
func isBlocked(statusCode int) bool {
	return statusCode == 403 || statusCode == 429
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (alternative proxy header)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// fetchServerIP retrieves the server's public IP address using ipinfo.io
func fetchServerIP() string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipinfo.io/ip")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting server IP: %v\n", err)
		return "unknown"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading server IP response: %v\n", err)
		return "unknown"
	}

	return strings.TrimSpace(string(body))
}

// getServerIP returns the cached server IP
func getServerIP() string {
	ipMutex.RLock()
	defer ipMutex.RUnlock()
	if cachedServerIP == "" {
		return "fetching..."
	}
	return cachedServerIP
}
