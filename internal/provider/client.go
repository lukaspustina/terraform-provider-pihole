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
	SessionID string `json:"session_id"`
	CSRFToken string `json:"csrf_token"`
}

type DNSRecord struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

type CNAMERecord struct {
	Domain string `json:"domain"`
	Target string `json:"target"`
}

func NewPiholeClient(baseURL, password string, config ClientConfig) (*PiholeClient, error) {
	client := &PiholeClient{
		BaseURL:  baseURL,
		Password: password,
		Config:   config,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
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

func (c *PiholeClient) authenticate() error {
	authReq := AuthRequest{Password: c.Password}
	jsonData, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/auth", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response as raw JSON to handle Pi-hole API v6 structure
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w, body: %s", err, string(body))
	}

	// Extract session info from Pi-hole API v6 response structure
	sessionData, exists := rawResp["session"]
	if !exists {
		return fmt.Errorf("no session data found in auth response: %s", string(body))
	}

	sessionMap, ok := sessionData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid session data format in auth response: %s", string(body))
	}

	// Extract session ID and CSRF token
	if sid, exists := sessionMap["sid"]; exists {
		if sidStr, ok := sid.(string); ok {
			c.SessionID = sidStr
		}
	}

	if csrf, exists := sessionMap["csrf"]; exists {
		if csrfStr, ok := csrf.(string); ok {
			c.CSRFToken = csrfStr
		}
	}

	if c.SessionID == "" {
		return fmt.Errorf("no session ID found in auth response: %s", string(body))
	}

	return nil
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

		req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("X-FTL-SID", c.SessionID)
		req.Header.Set("X-FTL-CSRF", c.CSRFToken)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
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
	// PUT /api/config/dns/hosts/192.168.2.19%20docker.k.pustina.internal
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
