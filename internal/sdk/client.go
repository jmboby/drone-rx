package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// LicenseField represents a single license field from the Replicated SDK.
type LicenseField struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	ValueType string      `json:"valueType"`
}

// LicenseEntitlement represents a single entitlement in the license info response.
type LicenseEntitlement struct {
	Title     string      `json:"title"`
	Value     interface{} `json:"value"`
	ValueType string      `json:"valueType"`
}

// LicenseInfo holds top-level license metadata.
type LicenseInfo struct {
	LicenseID    string                        `json:"licenseID"`
	ChannelName  string                        `json:"channelName"`
	LicenseType  string                        `json:"licenseType"`
	Entitlements map[string]LicenseEntitlement  `json:"entitlements"`
}

// IsExpired checks if the license has passed its expiration date.
func (l *LicenseInfo) IsExpired() bool {
	ea, ok := l.Entitlements["expires_at"]
	if !ok {
		return false
	}
	dateStr, ok := ea.Value.(string)
	if !ok || dateStr == "" {
		return false
	}
	expiry, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return false
	}
	return time.Now().After(expiry)
}

// ExpirationDate returns the expiration date string, or empty if none set.
func (l *LicenseInfo) ExpirationDate() string {
	ea, ok := l.Entitlements["expires_at"]
	if !ok {
		return ""
	}
	dateStr, _ := ea.Value.(string)
	return dateStr
}

// UpdateInfo describes an available application update.
type UpdateInfo struct {
	VersionLabel string `json:"versionLabel"`
	CreatedAt    string `json:"createdAt"`
	ReleaseNotes string `json:"releaseNotes"`
}

// Client is an HTTP client for the Replicated SDK sidecar.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a Client pointed at the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetLicenseField fetches a single license field by name.
func (c *Client) GetLicenseField(fieldName string) (*LicenseField, error) {
	url := fmt.Sprintf("%s/api/v1/license/fields/%s", c.baseURL, fieldName)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("sdk get license field: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sdk get license field: unexpected status %d", resp.StatusCode)
	}
	var field LicenseField
	if err := json.NewDecoder(resp.Body).Decode(&field); err != nil {
		return nil, fmt.Errorf("sdk get license field: decode: %w", err)
	}
	return &field, nil
}

// GetLicenseInfo fetches overall license information.
func (c *Client) GetLicenseInfo() (*LicenseInfo, error) {
	url := fmt.Sprintf("%s/api/v1/license/info", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("sdk get license info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sdk get license info: unexpected status %d", resp.StatusCode)
	}
	var info LicenseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("sdk get license info: decode: %w", err)
	}
	return &info, nil
}

// GetUpdates fetches available application updates.
func (c *Client) GetUpdates() ([]UpdateInfo, error) {
	url := fmt.Sprintf("%s/api/v1/app/updates", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("sdk get updates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sdk get updates: unexpected status %d", resp.StatusCode)
	}
	var updates []UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
		return nil, fmt.Errorf("sdk get updates: decode: %w", err)
	}
	return updates, nil
}

// SendMetrics posts custom metrics to the SDK. It logs errors but always returns nil (best-effort).
func (c *Client) SendMetrics(data map[string]interface{}) error {
	body := map[string]interface{}{"data": data}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Printf("sdk: metrics marshal error: %v", err)
		return nil
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v1/app/custom-metrics", c.baseURL),
		"application/json",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		log.Printf("sdk: metrics send failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("sdk: metrics send returned %d", resp.StatusCode)
	}
	return nil
}

// IsFeatureEnabled returns true if the named license field is truthy.
// Handles both boolean (true) and string ("true", "1") values from the SDK.
// Returns false on any error (fail closed).
func (c *Client) IsFeatureEnabled(fieldName string) bool {
	field, err := c.GetLicenseField(fieldName)
	if err != nil {
		log.Printf("sdk: feature check %s failed: %v", fieldName, err)
		return false
	}
	switch v := field.Value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	case float64:
		return v == 1
	default:
		return false
	}
}
