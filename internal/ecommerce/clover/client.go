package clover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CloverClient handles Clover API interactions
type CloverClient struct {
	BaseURL      string // Regular API (e.g., https://sandbox.dev.clover.com)
	EcommerceURL string // Ecommerce API (e.g., https://scl-sandbox.dev.clover.com)
	merchantID   string
	PrivateKey   string
	publicKey    string
	httpClient   *http.Client
}

// NewCloverClient creates a new Clover client from the provided values.
func NewCloverClient(baseURL, ecomURL, merchantID, privateKey, publicKey string) *CloverClient {
	return &CloverClient{
		BaseURL:      baseURL,
		EcommerceURL: ecomURL,
		merchantID:   merchantID,
		PrivateKey:   privateKey,
		publicKey:    publicKey,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetPublicKey returns the public API token for the frontend SDK.
func (c *CloverClient) GetPublicKey() string { return c.publicKey }

// GetMerchantID returns the merchant ID.
func (c *CloverClient) GetMerchantID() string { return c.merchantID }

func (c *CloverClient) doRequest(method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("request encoding error: %w", err)
		}
		reqBody = bytes.NewBuffer(payload)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, path), reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("request creation error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.PrivateKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("response read error: %w", err)
	}

	return respBody, resp.StatusCode, nil
}
