package clover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CloverClient handles Clover API interactions
type CloverClient struct {
	BaseURL      string // Regular API (e.g., https://sandbox.dev.clover.com)
	EcommerceURL string // Ecommerce API (e.g., https://scl-sandbox.dev.clover.com)
	MerchantID   string
	PrivateKey   string
	PublicKey    string
	httpClient   *http.Client
}

// NewCloverClient creates a new Clover client
func NewCloverClient() *CloverClient {
	return &CloverClient{
		BaseURL:      os.Getenv("CLOVER_BASE_URL"),
		EcommerceURL: os.Getenv("CLOVER_ECOMMERCE_URL"),
		MerchantID:   os.Getenv("CLOVER_MERCHANT_ID"),
		PrivateKey:   os.Getenv("CLOVER_PRIVATE_API_TOKEN"),
		PublicKey:    os.Getenv("CLOVER_PUBLIC_API_TOKEN"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *CloverClient) doRequest(method, path string, body any) ([]byte, int, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("request encoding error: %w", err)
		}
		reqBody = bytes.NewBuffer(payload)
	}

	req, err := http.NewRequest(method, url, reqBody)
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

// ============================================================================
// Inventory API
// ============================================================================

// Item represents a Clover inventory item
type Item struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Price          int64  `json:"price"` // in cents
	PriceType      string `json:"priceType"`
	SKU            string `json:"sku"`
	Code           string `json:"code"`
	Available      bool   `json:"available"`
	Hidden         bool   `json:"hidden"`
	StockCount     int    `json:"stockCount"`
	ModifiedTime   int64  `json:"modifiedTime"`
	DefaultTaxRate bool   `json:"defaultTaxRates"`
}

// ItemsResponse represents the response from listing items
type ItemsResponse struct {
	Elements []Item `json:"elements"`
}

// ListItems fetches all items from Clover inventory
func (c *CloverClient) ListItems() ([]Item, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/items?expand=stockCount", c.MerchantID)
	body, status, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("clover API error: status %d", status)
	}

	var resp ItemsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return resp.Elements, nil
}

// GetItem fetches a single item by ID
func (c *CloverClient) GetItem(itemID string) (*Item, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/items/%s?expand=stockCount", c.MerchantID, itemID)
	body, status, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("clover API error: status %d", status)
	}

	var item Item
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &item, nil
}

// ============================================================================
// Orders API
// ============================================================================

// OrderLineItem represents a line item in an order
type OrderLineItem struct {
	ID       string `json:"id,omitempty"`
	ItemID   string `json:"item_id,omitempty"`
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity,omitempty"`
}

// Order represents a Clover order
type Order struct {
	ID           string          `json:"id,omitempty"`
	Currency     string          `json:"currency"`
	Total        int64           `json:"total"`
	State        string          `json:"state,omitempty"`
	Title        string          `json:"title,omitempty"`
	Note         string          `json:"note,omitempty"`
	LineItems    []OrderLineItem `json:"lineItems,omitempty"`
	CreatedTime  int64           `json:"createdTime,omitempty"`
	ModifiedTime int64           `json:"modifiedTime,omitempty"`
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	Currency  string          `json:"currency"`
	Title     string          `json:"title,omitempty"`
	Note      string          `json:"note,omitempty"`
	LineItems []OrderLineItem `json:"lineItems,omitempty"`
}

// OrdersResponse represents the response from listing orders
type OrdersResponse struct {
	Elements []Order `json:"elements"`
}

// CreateOrder creates a new order in Clover
func (c *CloverClient) CreateOrder(req CreateOrderRequest) (*Order, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders", c.MerchantID)
	body, status, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("clover API error: status %d, body: %s", status, string(body))
	}

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &order, nil
}

// GetOrder fetches a single order by ID
func (c *CloverClient) GetOrder(orderID string) (*Order, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders/%s?expand=lineItems", c.MerchantID, orderID)
	body, status, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("clover API error: status %d", status)
	}

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &order, nil
}

// ListOrders fetches all orders
func (c *CloverClient) ListOrders() ([]Order, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders?expand=lineItems", c.MerchantID)
	body, status, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("clover API error: status %d", status)
	}

	var resp OrdersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return resp.Elements, nil
}

// AddLineItemToOrder adds a line item to an existing order
func (c *CloverClient) AddLineItemToOrder(orderID string, item OrderLineItem) (*OrderLineItem, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders/%s/line_items", c.MerchantID, orderID)
	body, status, err := c.doRequest("POST", path, item)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("clover API error: status %d", status)
	}

	var lineItem OrderLineItem
	if err := json.Unmarshal(body, &lineItem); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &lineItem, nil
}

// ============================================================================
// Payment API (Ecommerce)
// ============================================================================

// ChargeRequest represents a Clover ecommerce charge request
type ChargeRequest struct {
	Amount   int    `json:"amount"`   // in cents
	Currency string `json:"currency"` // "usd"
	Source   string `json:"source"`   // payment token from iframe
	Ecomind  string `json:"ecomind"`  // "ecom" for ecommerce transactions
}

// ChargeResponse represents a Clover ecommerce charge response
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

// Keeping old types for backwards compatibility
type PaymentRequest = ChargeRequest
type PaymentResponse = ChargeResponse

// ProcessPayment processes a payment through Clover Ecommerce API
func (c *CloverClient) ProcessPayment(req ChargeRequest) (*ChargeResponse, error) {
	if c.PrivateKey == "" {
		return nil, fmt.Errorf("missing Clover private key - check CLOVER_PRIVATE_API_TOKEN")
	}

	if c.EcommerceURL == "" {
		return nil, fmt.Errorf("missing Clover ecommerce URL - check CLOVER_ECOMMERCE_URL")
	}

	if req.Amount <= 0 || req.Amount > 999999 {
		return nil, fmt.Errorf("invalid amount")
	}

	if req.Source == "" {
		return nil, fmt.Errorf("payment token required")
	}

	// Set defaults
	if req.Currency == "" {
		req.Currency = "usd"
	}
	if req.Ecomind == "" {
		req.Ecomind = "ecom"
	}

	// Build request to ecommerce API
	url := fmt.Sprintf("%s/v1/charges", c.EcommerceURL)

	payload, err := json.Marshal(req)
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

	// Log raw response for debugging
	fmt.Printf("Clover charge response (status %d): %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("clover API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var chargeResp ChargeResponse
	if err := json.Unmarshal(body, &chargeResp); err != nil {
		return nil, fmt.Errorf("response parsing error: %w, body: %s", err, string(body))
	}

	return &chargeResp, nil
}

// ============================================================================
// Refunds API
// ============================================================================

// RefundRequest represents a Clover refund request
type RefundRequest struct {
	PaymentID string `json:"paymentId"`
	Amount    int64  `json:"amount,omitempty"` // partial refund amount in cents, omit for full refund
	Reason    string `json:"reason,omitempty"`
}

// RefundResponse represents a Clover refund response
type RefundResponse struct {
	ID        string `json:"id"`
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdTime,omitempty"`
}

// CreateRefund processes a refund through Clover
func (c *CloverClient) CreateRefund(req RefundRequest) (*RefundResponse, error) {
	if c.PrivateKey == "" || c.MerchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	if req.PaymentID == "" {
		return nil, fmt.Errorf("payment ID required")
	}

	path := fmt.Sprintf("/v3/merchants/%s/refunds", c.MerchantID)
	body, status, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var refundResp RefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return &refundResp, fmt.Errorf("clover API error: %d", status)
	}

	return &refundResp, nil
}

