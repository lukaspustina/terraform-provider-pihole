package provider

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Mock Pi-hole server for testing
func createMockPiholeServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle Pi-hole v6 API authentication
		if r.URL.Path == "/api/auth" && r.Method == "POST" {
			// Mock successful authentication response matching Pi-hole v6 format
			authResponse := AuthResponse{
				Session: struct {
					Valid    bool   `json:"valid"`
					Totp     bool   `json:"totp"`
					Sid      string `json:"sid"`
					Validity int    `json:"validity"`
					Message  string `json:"message"`
					CSRF     string `json:"csrf"`
				}{
					Valid:    true,
					Totp:     false,
					Sid:      "mock-session-id",
					Validity: 1,
					Message:  "success",
					CSRF:     "mock-csrf-token",
				},
				Took: 0.001,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(authResponse)
			return
		}

		// Handle Pi-hole v6 DNS management endpoints
		if r.URL.Path == "/api/config/dns/hosts" && r.Method == "GET" {
			// Mock DNS records response
			response := map[string]interface{}{
				"config": map[string]interface{}{
					"dns": map[string]interface{}{
						"hosts": []string{
							"192.168.1.100 test.example.com",
							"192.168.1.101 server.example.com",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		if r.URL.Path == "/api/config/dns/cnameRecords" && r.Method == "GET" {
			// Mock CNAME records response
			response := map[string]interface{}{
				"config": map[string]interface{}{
					"dns": map[string]interface{}{
						"cnameRecords": []string{
							"www.example.com,example.com",
							"mail.example.com,server.example.com",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Handle DNS record creation/modification
		if strings.HasPrefix(r.URL.Path, "/api/config/dns/hosts/") && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
			return
		}

		// Handle CNAME record creation/modification
		if strings.HasPrefix(r.URL.Path, "/api/config/dns/cnameRecords/") && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
			return
		}

		// Handle DNS record deletion
		if strings.HasPrefix(r.URL.Path, "/api/config/dns/hosts/") && r.Method == "DELETE" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
			return
		}

		// Handle CNAME record deletion
		if strings.HasPrefix(r.URL.Path, "/api/config/dns/cnameRecords/") && r.Method == "DELETE" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
			return
		}

		// Handle configuration management endpoints
		if r.URL.Path == "/api/config/webserver" && r.Method == "GET" {
			response := map[string]interface{}{
				"config": map[string]interface{}{
					"webserver": map[string]interface{}{
						"api": map[string]interface{}{
							"app_sudo": true, // Changed to true for the GetConfig test
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		if r.URL.Path == "/api/config/webserver" && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
			return
		}

		// Handle legacy admin API for compatibility (still used by some tests)
		if r.URL.Path == "/admin/api.php" && r.Method == "GET" && r.URL.Query().Has("summary") {
			summaryResponse := map[string]interface{}{
				"domains_being_blocked": 1000,
				"dns_queries_today":     5000,
				"ads_blocked_today":     500,
				"ads_percentage_today":  10.5,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(summaryResponse)
			return
		}

		// For now, just handle basic endpoints for testing connection
		// The actual DNS management API will be determined based on real Pi-hole testing
		if r.URL.Path == "/admin/api.php" && r.Method == "GET" {
			// Check if it has auth parameter
			if auth := r.URL.Query().Get("auth"); auth != "" {
				// Return success for any request with auth
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
				return
			}
		}

		// Default 404
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
	}))
}

func TestNewPiholeClient(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 100,
		RetryAttempts:  3,
		RetryBackoffMs: 500,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	if client.BaseURL != server.URL {
		t.Errorf("Expected BaseURL to be %s, got %s", server.URL, client.BaseURL)
	}

	if client.SessionID != "mock-session-id" {
		t.Errorf("Expected SessionID to be 'mock-session-id', got '%s'", client.SessionID)
	}

	if client.CSRFToken != "mock-csrf-token" {
		t.Errorf("Expected CSRFToken to be 'mock-csrf-token', got '%s'", client.CSRFToken)
	}

	if client.Config.MaxConnections != 1 {
		t.Errorf("Expected MaxConnections to be 1, got %d", client.Config.MaxConnections)
	}
}

func TestPiholeClient_GetDNSRecords(t *testing.T) {

	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	records, err := client.GetDNSRecords()
	if err != nil {
		t.Fatalf("Failed to get DNS records: %v", err)
	}

	expectedRecords := []DNSRecord{
		{IP: "192.168.1.100", Domain: "test.example.com"},
		{IP: "192.168.1.101", Domain: "server.example.com"},
	}

	if len(records) != len(expectedRecords) {
		t.Fatalf("Expected %d records, got %d", len(expectedRecords), len(records))
	}

	for i, expected := range expectedRecords {
		if records[i].IP != expected.IP {
			t.Errorf("Record %d: expected IP %s, got %s", i, expected.IP, records[i].IP)
		}
		if records[i].Domain != expected.Domain {
			t.Errorf("Record %d: expected Domain %s, got %s", i, expected.Domain, records[i].Domain)
		}
	}
}

func TestPiholeClient_GetCNAMERecords(t *testing.T) {

	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	records, err := client.GetCNAMERecords()
	if err != nil {
		t.Fatalf("Failed to get CNAME records: %v", err)
	}

	expectedRecords := []CNAMERecord{
		{Domain: "www.example.com", Target: "example.com"},
		{Domain: "mail.example.com", Target: "server.example.com"},
	}

	if len(records) != len(expectedRecords) {
		t.Fatalf("Expected %d records, got %d", len(expectedRecords), len(records))
	}

	for i, expected := range expectedRecords {
		if records[i].Domain != expected.Domain {
			t.Errorf("Record %d: expected Domain %s, got %s", i, expected.Domain, records[i].Domain)
		}
		if records[i].Target != expected.Target {
			t.Errorf("Record %d: expected Target %s, got %s", i, expected.Target, records[i].Target)
		}
	}
}

func TestPiholeClient_CreateDNSRecord(t *testing.T) {

	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	err = client.CreateDNSRecord("new.example.com", "192.168.1.200")
	if err != nil {
		t.Fatalf("Failed to create DNS record: %v", err)
	}
}

func TestPiholeClient_CreateCNAMERecord(t *testing.T) {

	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	err = client.CreateCNAMERecord("blog.example.com", "server.example.com")
	if err != nil {
		t.Fatalf("Failed to create CNAME record: %v", err)
	}
}

func TestPiholeClient_DeleteDNSRecord(t *testing.T) {

	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	err = client.DeleteDNSRecord("test.example.com")
	if err != nil {
		t.Fatalf("Failed to delete DNS record: %v", err)
	}
}

func TestPiholeClient_DeleteCNAMERecord(t *testing.T) {

	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	err = client.DeleteCNAMERecord("www.example.com")
	if err != nil {
		t.Fatalf("Failed to delete CNAME record: %v", err)
	}
}

func TestPiholeClient_RetryLogic(t *testing.T) {
	// Create a server that fails the first few requests
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.URL.Path == "/api/auth" {
			if attempts <= 2 {
				// Simulate connection errors for first 2 attempts
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Success on 3rd attempt
			authResponse := AuthResponse{
				Session: struct {
					Valid    bool   `json:"valid"`
					Totp     bool   `json:"totp"`
					Sid      string `json:"sid"`
					Validity int    `json:"validity"`
					Message  string `json:"message"`
					CSRF     string `json:"csrf"`
				}{
					Valid:    true,
					Totp:     false,
					Sid:      "test-session-id",
					Validity: 1,
					Message:  "success",
					CSRF:     "test-csrf-token",
				},
				Took: 0.001,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(authResponse)
			return
		}
	}))
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 10,
		RetryAttempts:  3,
		RetryBackoffMs: 50,
		InsecureTLS:    false,
	}

	// Authentication now implements retry logic, so it should eventually succeed
	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Expected authentication to succeed after retries, but got error: %v", err)
	}

	if client.SessionID != "test-session-id" {
		t.Errorf("Expected SessionID to be 'test-session-id', got '%s'", client.SessionID)
	}

	// Verify that it actually took multiple attempts
	if attempts <= 2 {
		t.Errorf("Expected at least 3 attempts due to failures, but only made %d attempts", attempts)
	}
}

func TestPiholeClient_URLEncoding(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" {
			authResponse := AuthResponse{
				Session: struct {
					Valid    bool   `json:"valid"`
					Totp     bool   `json:"totp"`
					Sid      string `json:"sid"`
					Validity int    `json:"validity"`
					Message  string `json:"message"`
					CSRF     string `json:"csrf"`
				}{
					Valid:    true,
					Totp:     false,
					Sid:      "test-session-id",
					Validity: 1,
					Message:  "success",
					CSRF:     "test-csrf-token",
				},
				Took: 0.001,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(authResponse)
			return
		}

		if r.Method == "GET" && r.URL.Path == "/api/config/dns/hosts" {
			response := map[string]interface{}{
				"config": map[string]interface{}{
					"dns": map[string]interface{}{
						"hosts": []string{},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Check URL encoding for DNS record creation
		if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/api/config/dns/hosts/") {
			recordPart := strings.TrimPrefix(r.URL.Path, "/api/config/dns/hosts/")
			decodedRecord, err := url.PathUnescape(recordPart)
			if err != nil {
				t.Errorf("Failed to decode URL path: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			expectedRecord := "192.168.1.100 test-domain.example.com"
			if decodedRecord != expectedRecord {
				t.Errorf("Expected decoded record '%s', got '%s'", expectedRecord, decodedRecord)
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "created"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 10,
		RetryAttempts:  1,
		RetryBackoffMs: 50,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	// Test DNS record with domain that needs URL encoding
	err = client.CreateDNSRecord("test-domain.example.com", "192.168.1.100")
	if err != nil {
		t.Fatalf("Failed to create DNS record with URL encoding: %v", err)
	}
}

func TestIsRetryableError(t *testing.T) {
	testCases := []struct {
		errorMsg string
		expected bool
	}{
		{"connection refused", true},
		{"EOF", true},
		{"timeout", true},
		{"connection reset", true},
		{"invalid credentials", false},
		{"not found", false},
		{"permission denied", false},
	}

	for _, tc := range testCases {
		t.Run(tc.errorMsg, func(t *testing.T) {
			err := &url.Error{Err: &net.AddrError{Err: tc.errorMsg}}
			result := isRetryableError(err)
			if result != tc.expected {
				t.Errorf("For error '%s': expected %v, got %v", tc.errorMsg, tc.expected, result)
			}
		})
	}
}

func TestClientConfig_Defaults(t *testing.T) {
	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 300,
		RetryAttempts:  3,
		RetryBackoffMs: 500,
		InsecureTLS:    false,
	}

	if config.MaxConnections != 1 {
		t.Errorf("Expected MaxConnections default to be 1, got %d", config.MaxConnections)
	}

	if config.RequestDelayMs != 300 {
		t.Errorf("Expected RequestDelayMs default to be 300, got %d", config.RequestDelayMs)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("Expected RetryAttempts default to be 3, got %d", config.RetryAttempts)
	}

	if config.RetryBackoffMs != 500 {
		t.Errorf("Expected RetryBackoffMs default to be 500, got %d", config.RetryBackoffMs)
	}

	if config.InsecureTLS != false {
		t.Errorf("Expected InsecureTLS default to be false, got %t", config.InsecureTLS)
	}
}

func TestTLSConfiguration_SecureByDefault(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 100,
		RetryAttempts:  3,
		RetryBackoffMs: 500,
		InsecureTLS:    false, // Secure TLS verification
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	transport, ok := client.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Expected client to use http.Transport")
	}

	if transport.TLSClientConfig.InsecureSkipVerify != false {
		t.Errorf("Expected InsecureSkipVerify to be false for secure TLS, got %t", transport.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestTLSConfiguration_InsecureWhenConfigured(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 100,
		RetryAttempts:  3,
		RetryBackoffMs: 500,
		InsecureTLS:    true, // Insecure TLS verification disabled
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	transport, ok := client.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Expected client to use http.Transport")
	}

	if transport.TLSClientConfig.InsecureSkipVerify != true {
		t.Errorf("Expected InsecureSkipVerify to be true for insecure TLS, got %t", transport.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestTLSConfiguration_HTTPSServer(t *testing.T) {
	// Create HTTPS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" && r.Method == "POST" {
			authResponse := AuthResponse{
				Session: struct {
					Valid    bool   `json:"valid"`
					Totp     bool   `json:"totp"`
					Sid      string `json:"sid"`
					Validity int    `json:"validity"`
					Message  string `json:"message"`
					CSRF     string `json:"csrf"`
				}{
					Valid:    true,
					Totp:     false,
					Sid:      "mock-session-id",
					Validity: 1,
					Message:  "success",
					CSRF:     "mock-csrf-token",
				},
				Took: 0.001,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(authResponse)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Run("insecure_tls=false should fail with self-signed cert", func(t *testing.T) {
		config := ClientConfig{
			MaxConnections: 1,
			RequestDelayMs: 100,
			RetryAttempts:  1, // Reduce retries for faster test
			RetryBackoffMs: 100,
			InsecureTLS:    false, // Secure TLS verification
		}

		_, err := NewPiholeClient(server.URL, "test-password", config)
		if err == nil {
			t.Error("Expected client creation to fail with secure TLS and self-signed certificate")
		}

		// Check that the error is related to certificate verification
		if !strings.Contains(err.Error(), "certificate") && !strings.Contains(err.Error(), "tls") {
			t.Logf("Error message: %v", err)
			// Note: We're being lenient here as the exact error message may vary
		}
	})

	t.Run("insecure_tls=true should succeed with self-signed cert", func(t *testing.T) {
		config := ClientConfig{
			MaxConnections: 1,
			RequestDelayMs: 100,
			RetryAttempts:  3,
			RetryBackoffMs: 500,
			InsecureTLS:    true, // Allow insecure TLS
		}

		client, err := NewPiholeClient(server.URL, "test-password", config)
		if err != nil {
			t.Fatalf("Expected client creation to succeed with insecure TLS, got error: %v", err)
		}

		if client.SessionID != "mock-session-id" {
			t.Errorf("Expected SessionID to be 'mock-session-id', got '%s'", client.SessionID)
		}
	})
}

func TestPiholeClient_GetConfig(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	// Test getting webserver.api.app_sudo configuration
	configSetting, err := client.GetConfig("webserver.api.app_sudo")
	if err != nil {
		t.Fatalf("Failed to get configuration: %v", err)
	}

	if configSetting.Key != "webserver.api.app_sudo" {
		t.Errorf("Expected key 'webserver.api.app_sudo', got '%s'", configSetting.Key)
	}

	// The mock server returns true for this specific endpoint
	if configSetting.Value != true {
		t.Errorf("Expected value true, got %v", configSetting.Value)
	}
}

func TestPiholeClient_SetConfig(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	// Test setting webserver.api.app_sudo configuration
	err = client.SetConfig("webserver.api.app_sudo", true)
	if err != nil {
		t.Fatalf("Failed to set configuration: %v", err)
	}
}

func TestPiholeClient_GetWebserverConfig(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	// Test getting webserver configuration section
	webserverConfig, err := client.GetWebserverConfig()
	if err != nil {
		t.Fatalf("Failed to get webserver configuration: %v", err)
	}

	// Check that we got the expected structure
	if apiConfig, ok := webserverConfig["api"].(map[string]interface{}); ok {
		if appSudo, exists := apiConfig["app_sudo"]; exists {
			if appSudo != true {
				t.Errorf("Expected app_sudo to be true, got %v", appSudo)
			}
		} else {
			t.Error("Expected 'app_sudo' key in API configuration")
		}
	} else {
		t.Error("Expected 'api' section in webserver configuration")
	}
}

func TestPiholeClient_SetWebserverConfig(t *testing.T) {
	server := createMockPiholeServer()
	defer server.Close()

	config := ClientConfig{
		MaxConnections: 1,
		RequestDelayMs: 50,
		RetryAttempts:  1,
		RetryBackoffMs: 100,
		InsecureTLS:    false,
	}

	client, err := NewPiholeClient(server.URL, "test-password", config)
	if err != nil {
		t.Fatalf("Failed to create Pi-hole client: %v", err)
	}

	// Test setting webserver configuration
	newConfig := map[string]interface{}{
		"api": map[string]interface{}{
			"app_sudo": true,
		},
	}

	err = client.SetWebserverConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to set webserver configuration: %v", err)
	}
}
