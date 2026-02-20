package clover

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ItemRef references a Clover inventory item by ID.
type ItemRef struct {
	ID string `json:"id"`
}

// OrderLineItem represents a line item in an order
type OrderLineItem struct {
	ID      string   `json:"id,omitempty"`
	Item    *ItemRef `json:"item,omitempty"` // reference to inventory item
	Name    string   `json:"name,omitempty"`
	Price   int64    `json:"price,omitempty"`
	UnitQty int      `json:"unitQty,omitempty"`
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
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	body, status, err := c.doRequest("POST", fmt.Sprintf("/v3/merchants/%s/orders", c.merchantID), req)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("clover API error: status %d", status)
	}

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &order, nil
}

// GetOrder fetches a single order by ID
func (c *CloverClient) GetOrder(orderID string) (*Order, error) {
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders/%s?expand=lineItems", c.merchantID, orderID)
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
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders?expand=lineItems", c.merchantID)
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

// DeleteOrder deletes an order by ID.
func (c *CloverClient) DeleteOrder(orderID string) error {
	if c.PrivateKey == "" || c.merchantID == "" {
		return fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders/%s", c.merchantID, orderID)
	_, status, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("clover API error: status %d", status)
	}

	return nil
}

// AddLineItemToOrder adds a line item to an existing order
func (c *CloverClient) AddLineItemToOrder(orderID string, item OrderLineItem) (*OrderLineItem, error) {
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/orders/%s/line_items", c.merchantID, orderID)
	body, status, err := c.doRequest("POST", path, item)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("clover API error: status %d, body: %s", status, string(body))
	}

	var lineItem OrderLineItem
	if err := json.Unmarshal(body, &lineItem); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return &lineItem, nil
}
