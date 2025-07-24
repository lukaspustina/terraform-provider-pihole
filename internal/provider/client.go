package provider

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientConfig struct {
	MaxConnections int
	RequestDelayMs int
	RetryAttempts  int
	RetryBackoffMs int
	InsecureTLS    bool
}

type PiholeClient struct {
	BaseURL    string
	Password   string
	HTTPClient *http.Client
	SessionID  string
	CSRFToken  string
	Config     ClientConfig
}

type AuthRequest struct {
	Password string `json:"password"`
}

type AuthResponse struct {
	Session struct {
		Valid    bool   `json:"valid"`
		Totp     bool   `json:"totp"`
		Sid      string `json:"sid"`
		Validity int    `json:"validity"`
		Message  string `json:"message"`
		CSRF     string `json:"csrf"`
	} `json:"session"`
	Took float64 `json:"took"`
}

type DNSRecord struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

type CNAMERecord struct {
	Domain string `json:"domain"`
	Target string `json:"target"`
}

type ConfigSetting struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func NewPiholeClient(baseURL, password string, config ClientConfig) (*PiholeClient, error) {
	client := &PiholeClient{
		BaseURL:  baseURL,
		Password: password,
		Config:   config,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: config.InsecureTLS},
				DisableKeepAlives: false,
				IdleConnTimeout:   90 * time.Second,
				MaxIdleConns:      10,
				MaxConnsPerHost:   config.MaxConnections,
			},
		},
	}

	if err := client.authenticate(); err != nil {
		return nil, err
	}

	return client, nil
}

// Close cleans up the Pi-hole client session
func (c *PiholeClient) Close() error {
	// Pi-hole v6 sessions automatically expire, but we can clear our tokens
	c.SessionID = ""
	c.CSRFToken = ""
	return nil
}

func (c *PiholeClient) authenticate() error {
	return c.authenticateWithRetry(c.Config.RetryAttempts)
}

func (c *PiholeClient) authenticateWithRetry(retries int) error {
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		// Add delay between attempts (exponential backoff)
		if attempt > 0 {
			backoffDelay := time.Duration(attempt*attempt) * time.Duration(c.Config.RetryBackoffMs) * time.Millisecond
			time.Sleep(backoffDelay)
		}

		// Pi-hole v6 API authentication via /api/auth
		authReq := AuthRequest{Password: c.Password}

		jsonData, err := json.Marshal(authReq)
		if err != nil {
			return fmt.Errorf("failed to marshal auth request: %w", err)
		}

		authURL := fmt.Sprintf("%s/api/auth", c.BaseURL)
		req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create auth request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			// Check if it's a connection error that might benefit from retry
			if isRetryableError(err) && attempt < retries {
				continue
			}
			return fmt.Errorf("failed to authenticate with Pi-hole: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			if attempt < retries {
				continue
			}
			return fmt.Errorf("failed to read auth response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("authentication failed with status: %d, body: %s", resp.StatusCode, string(body))
			// Don't retry authentication failures (401, 429, etc.)
			if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusTooManyRequests {
				return lastErr
			}
			if attempt < retries {
				continue
			}
			return lastErr
		}

		var authResp AuthResponse
		if err := json.Unmarshal(body, &authResp); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal auth response: %w, body: %s", err, string(body))
			if attempt < retries {
				continue
			}
			return lastErr
		}

		// Check if authentication was successful
		if !authResp.Session.Valid {
			lastErr = fmt.Errorf("authentication failed: %s", authResp.Session.Message)
			// Don't retry invalid credentials
			return lastErr
		}

		c.SessionID = authResp.Session.Sid
		c.CSRFToken = authResp.Session.CSRF

		return nil
	}

	return fmt.Errorf("authentication failed after %d attempts: %w", retries+1, lastErr)
}

func (c *PiholeClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	return c.makeRequestWithRetry(method, endpoint, body, c.Config.RetryAttempts)
}

func (c *PiholeClient) makeRequestWithRetry(method, endpoint string, body interface{}, retries int) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		// Add delay between attempts (exponential backoff)
		if attempt > 0 {
			backoffDelay := time.Duration(attempt*attempt) * time.Duration(c.Config.RetryBackoffMs) * time.Millisecond
			time.Sleep(backoffDelay)
		}

		var reqBody io.Reader
		if body != nil {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}

		// Build full URL for Pi-hole v6 API
		fullURL := c.BaseURL + endpoint

		req, err := http.NewRequest(method, fullURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Add Pi-hole v6 API headers
		if c.SessionID != "" {
			req.Header.Set("X-FTL-SID", c.SessionID)
		}
		if c.CSRFToken != "" {
			req.Header.Set("X-FTL-CSRF", c.CSRFToken)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			// Check if it's a connection error that might benefit from retry
			if isRetryableError(err) && attempt < retries {
				continue
			}
			return nil, err
		}

		// Success or non-retryable error
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", retries+1, lastErr)
}

func isRetryableError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection reset")
}

func (c *PiholeClient) GetDNSRecords() ([]DNSRecord, error) {
	resp, err := c.makeRequest("GET", "/api/config/dns/hosts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DNS records response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get DNS records, status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse Pi-hole API v6 response structure
	var apiResp struct {
		Config struct {
			DNS struct {
				Hosts []string `json:"hosts"`
			} `json:"dns"`
		} `json:"config"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNS records: %w, body: %s", err, string(body))
	}

	var records []DNSRecord
	for _, recordStr := range apiResp.Config.DNS.Hosts {
		parts := strings.SplitN(recordStr, " ", 2)
		if len(parts) == 2 {
			records = append(records, DNSRecord{
				IP:     parts[0],
				Domain: parts[1],
			})
		}
	}

	return records, nil
}

func (c *PiholeClient) CreateDNSRecord(domain, ip string) error {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	// Check if record already exists
	currentRecords, err := c.GetDNSRecords()
	if err != nil {
		return fmt.Errorf("failed to get current DNS records: %w", err)
	}

	for _, record := range currentRecords {
		if record.Domain == domain {
			if record.IP != ip {
				// Update existing record
				return c.UpdateDNSRecord(domain, ip)
			}
			// Record already exists with same IP, nothing to do
			return nil
		}
	}

	// Pi-hole API v6 format: everything in URL with URL-encoded space
	// PUT /api/config/dns/hosts/192.168.0.22%20www.homelab.local
	recordValue := fmt.Sprintf("%s %s", ip, domain)
	encodedRecord := url.PathEscape(recordValue)
	endpoint := fmt.Sprintf("/api/config/dns/hosts/%s", encodedRecord)

	resp, err := c.makeRequest("PUT", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("failed to create DNS record at %s, status: %d, body: %s", endpoint, resp.StatusCode, string(body))
}

func (c *PiholeClient) UpdateDNSRecord(domain, ip string) error {
	// First delete the old record, then create the new one
	if err := c.DeleteDNSRecord(domain); err != nil {
		return fmt.Errorf("failed to delete old DNS record: %w", err)
	}

	// Now create the new record
	return c.CreateDNSRecord(domain, ip)
}

func (c *PiholeClient) DeleteDNSRecord(domain string) error {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	// Get current records to find the exact record to delete
	currentRecords, err := c.GetDNSRecords()
	if err != nil {
		return fmt.Errorf("failed to get current DNS records: %w", err)
	}

	// Find the record to delete
	var recordToDelete *DNSRecord
	for _, record := range currentRecords {
		if record.Domain == domain {
			recordToDelete = &record
			break
		}
	}

	if recordToDelete == nil {
		// Record doesn't exist, consider it already deleted
		return nil
	}

	// Use DELETE method with URL-encoded record value in path
	recordValue := fmt.Sprintf("%s %s", recordToDelete.IP, recordToDelete.Domain)
	encodedRecord := url.PathEscape(recordValue)
	endpoint := fmt.Sprintf("/api/config/dns/hosts/%s", encodedRecord)

	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	return fmt.Errorf("failed to delete DNS record, status: %d, body: %s", resp.StatusCode, string(body))
}

func (c *PiholeClient) GetCNAMERecords() ([]CNAMERecord, error) {
	resp, err := c.makeRequest("GET", "/api/config/dns/cnameRecords", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get CNAME records: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CNAME records response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get CNAME records, status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse Pi-hole API v6 response structure
	var apiResp struct {
		Config struct {
			DNS struct {
				CNAMERecords []string `json:"cnameRecords"`
			} `json:"dns"`
		} `json:"config"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CNAME records: %w, body: %s", err, string(body))
	}

	var records []CNAMERecord
	for _, recordStr := range apiResp.Config.DNS.CNAMERecords {
		parts := strings.SplitN(recordStr, ",", 2)
		if len(parts) == 2 {
			records = append(records, CNAMERecord{
				Domain: parts[0],
				Target: parts[1],
			})
		}
	}

	return records, nil
}

func (c *PiholeClient) CreateCNAMERecord(domain, target string) error {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	// Check if record already exists
	currentRecords, err := c.GetCNAMERecords()
	if err != nil {
		return fmt.Errorf("failed to get current CNAME records: %w", err)
	}

	for _, record := range currentRecords {
		if record.Domain == domain {
			if record.Target != target {
				// Update existing record
				return c.UpdateCNAMERecord(domain, target)
			}
			// Record already exists with same target, nothing to do
			return nil
		}
	}

	// Pi-hole API v6 format: everything in URL with comma separator
	// PUT /api/config/dns/cnameRecords/www.example.com,example.com
	recordValue := fmt.Sprintf("%s,%s", domain, target)
	encodedRecord := url.PathEscape(recordValue)
	endpoint := fmt.Sprintf("/api/config/dns/cnameRecords/%s", encodedRecord)

	resp, err := c.makeRequest("PUT", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create CNAME record: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("failed to create CNAME record at %s, status: %d, body: %s", endpoint, resp.StatusCode, string(body))
}

func (c *PiholeClient) UpdateCNAMERecord(domain, target string) error {
	// First delete the old record, then create the new one
	if err := c.DeleteCNAMERecord(domain); err != nil {
		return fmt.Errorf("failed to delete old CNAME record: %w", err)
	}

	// Now create the new record
	return c.CreateCNAMERecord(domain, target)
}

func (c *PiholeClient) DeleteCNAMERecord(domain string) error {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	// Get current records to find the exact record to delete
	currentRecords, err := c.GetCNAMERecords()
	if err != nil {
		return fmt.Errorf("failed to get current CNAME records: %w", err)
	}

	// Find the record to delete
	var recordToDelete *CNAMERecord
	for _, record := range currentRecords {
		if record.Domain == domain {
			recordToDelete = &record
			break
		}
	}

	if recordToDelete == nil {
		// Record doesn't exist, consider it already deleted
		return nil
	}

	// Use DELETE method with URL-encoded record value in path
	recordValue := fmt.Sprintf("%s,%s", recordToDelete.Domain, recordToDelete.Target)
	encodedRecord := url.PathEscape(recordValue)
	endpoint := fmt.Sprintf("/api/config/dns/cnameRecords/%s", encodedRecord)

	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete CNAME record: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	return fmt.Errorf("failed to delete CNAME record, status: %d, body: %s", resp.StatusCode, string(body))
}

// GetConfig retrieves a specific configuration setting from Pi-hole
func (c *PiholeClient) GetConfig(configKey string) (*ConfigSetting, error) {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	// Determine the appropriate endpoint based on the configuration key
	var endpoint string
	configParts := strings.Split(configKey, ".")

	// For webserver configuration keys, use the webserver endpoint
	if len(configParts) > 0 && configParts[0] == "webserver" {
		endpoint = "/api/config/webserver"
	} else {
		// For other configurations, use a general endpoint
		endpoint = fmt.Sprintf("/api/config/%s", configParts[0])
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration '%s': %w", configKey, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get configuration '%s', status: %d, body: %s", configKey, resp.StatusCode, string(body))
	}

	// Parse the response - expecting nested config structure
	var apiResp struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration response: %w, body: %s", err, string(body))
	}

	// Navigate through the nested configuration structure
	// For webserver.api.app_sudo, we need to navigate: config.webserver.api.app_sudo
	var currentValue interface{} = apiResp.Config

	for _, part := range configParts {
		if configMap, ok := currentValue.(map[string]interface{}); ok {
			if val, exists := configMap[part]; exists {
				currentValue = val
			} else {
				return nil, fmt.Errorf("configuration key '%s' not found", configKey)
			}
		} else {
			return nil, fmt.Errorf("configuration structure is not as expected for key '%s'", configKey)
		}
	}

	return &ConfigSetting{
		Key:   configKey,
		Value: currentValue,
	}, nil
}

// SetConfig updates a specific configuration setting in Pi-hole
func (c *PiholeClient) SetConfig(configKey string, value interface{}) error {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	configParts := strings.Split(configKey, ".")

	// Handle webserver configuration specially
	if len(configParts) > 0 && configParts[0] == "webserver" {
		return c.setWebserverConfigValue(configKey, value)
	}

	// For other configurations, use a more general approach
	// This is a placeholder for future configuration types
	return fmt.Errorf("configuration type '%s' not yet supported", configParts[0])
}

// setWebserverConfigValue updates a webserver configuration value
func (c *PiholeClient) setWebserverConfigValue(configKey string, value interface{}) error {
	// First get the current webserver configuration
	currentConfig, err := c.GetWebserverConfig()
	if err != nil {
		return fmt.Errorf("failed to get current webserver config: %w", err)
	}

	// Parse the configuration key (e.g., "webserver.api.app_sudo")
	configParts := strings.Split(configKey, ".")
	if len(configParts) < 2 {
		return fmt.Errorf("invalid webserver configuration key format: %s", configKey)
	}

	// Skip the "webserver" part and navigate the rest
	keyParts := configParts[1:]

	// Create a deep copy of current config and update the specific value
	updatedConfig := make(map[string]interface{})
	for k, v := range currentConfig {
		updatedConfig[k] = v
	}

	// Navigate and update the nested structure
	current := updatedConfig
	for i, part := range keyParts {
		if i == len(keyParts)-1 {
			// Last part - set the value
			current[part] = value
		} else {
			// Intermediate part - ensure the nested map exists
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			}
			if nested, ok := current[part].(map[string]interface{}); ok {
				current = nested
			} else {
				return fmt.Errorf("configuration path '%s' is not a nested object", strings.Join(keyParts[:i+1], "."))
			}
		}
	}

	// Update the webserver configuration
	return c.SetWebserverConfig(updatedConfig)
}

// GetWebserverConfig retrieves the webserver configuration section
func (c *PiholeClient) GetWebserverConfig() (map[string]interface{}, error) {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	resp, err := c.makeRequest("GET", "/api/config/webserver", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get webserver configuration: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read webserver configuration response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get webserver configuration, status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var apiResp struct {
		Config struct {
			Webserver map[string]interface{} `json:"webserver"`
		} `json:"config"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webserver configuration: %w, body: %s", err, string(body))
	}

	return apiResp.Config.Webserver, nil
}

// SetWebserverConfig updates webserver configuration settings
func (c *PiholeClient) SetWebserverConfig(config map[string]interface{}) error {
	// Add delay to prevent overwhelming the API
	time.Sleep(time.Duration(c.Config.RequestDelayMs) * time.Millisecond)

	resp, err := c.makeRequest("PUT", "/api/config/webserver", config)
	if err != nil {
		return fmt.Errorf("failed to set webserver configuration: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	return fmt.Errorf("failed to set webserver configuration, status: %d, body: %s", resp.StatusCode, string(body))
}
