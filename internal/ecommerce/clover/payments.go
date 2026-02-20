package clover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ChargeResponse represents a Clover ecommerce payment response.
type ChargeResponse struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Paid     bool   `json:"paid"`
	Captured bool   `json:"captured"`
	RefNum   string `json:"ref_num"`
	AuthCode string `json:"auth_code"`
}

// PayOrderRequest represents a request to pay for an existing order.
type PayOrderRequest struct {
	Source       string `json:"source"`                  // payment token from iframe
	Ecomind      string `json:"ecomind"`                 // "ecom" for ecommerce transactions
	ReceiptEmail string `json:"receipt_email,omitempty"` // email for Clover receipt (production only)
}

// PayForOrder pays for an existing order, linking payment to line items.
func (c *CloverClient) PayForOrder(orderID string, source string, receiptEmail string) (*ChargeResponse, error) {
	if c.PrivateKey == "" {
		return nil, fmt.Errorf("missing Clover private key - check CLOVER_PRIVATE_API_TOKEN")
	}
	if c.EcommerceURL == "" {
		return nil, fmt.Errorf("missing Clover ecommerce URL - check CLOVER_ECOMMERCE_URL")
	}
	if source == "" {
		return nil, fmt.Errorf("payment token required")
	}

	reqBody := PayOrderRequest{
		Source:       source,
		Ecomind:      "ecom",
		ReceiptEmail: receiptEmail,
	}

	url := fmt.Sprintf("%s/v1/orders/%s/pay", c.EcommerceURL, orderID)

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("request encoding error: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.PrivateKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response read error: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("clover API error: status %d", resp.StatusCode)
	}

	var chargeResp ChargeResponse
	if err := json.Unmarshal(body, &chargeResp); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &chargeResp, nil
}
