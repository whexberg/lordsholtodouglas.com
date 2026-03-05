package clover

import (
	"encoding/json"
	"fmt"
	"lsd3/internal/ecommerce/cart"
	"net/http"
)

// ItemStock represents the nested stock data from Clover's itemStock expansion.
type ItemStock struct {
	Quantity float64 `json:"quantity"`
}

// Item represents a Clover inventory item
type Item struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Price            int64      `json:"price"` // in cents
	PriceType        string     `json:"priceType"`
	SKU              string     `json:"sku"`
	Code             string     `json:"code"`
	Available        bool       `json:"available"`
	Hidden           bool       `json:"hidden"`
	AutoManageStock  bool       `json:"autoManage"`
	Stock            *ItemStock `json:"itemStock,omitempty"`
	ModifiedTime     int64      `json:"modifiedTime"`
	DefaultTaxRate   bool       `json:"defaultTaxRates"`
}

// StockCount returns the item's stock quantity as an int.
func (i *Item) StockCount() int {
	if i.Stock == nil {
		return 0
	}
	return int(i.Stock.Quantity)
}

// ItemsResponse represents the response from listing items
type ItemsResponse struct {
	Elements []Item `json:"elements"`
}

// ListItems fetches all items from Clover inventory
func (c *CloverClient) ListItems() ([]Item, error) {
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/items?expand=itemStock", c.merchantID)
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

// LookupItem returns authoritative item data for the cart.
// This satisfies the cart.ItemLookup interface.
func (c *CloverClient) LookupItem(itemID string) (*cart.ItemInfo, error) {
	item, err := c.GetItem(itemID)
	if err != nil {
		return nil, err
	}
	if item.Hidden || !item.Available {
		return nil, fmt.Errorf("item %s is not available", itemID)
	}
	stock := -1 // unlimited
	if item.AutoManageStock {
		stock = item.StockCount()
	}
	return &cart.ItemInfo{
		Name:          item.Name,
		PriceCents:    item.Price,
		StockCount:    stock,
		VariablePrice: item.PriceType == "VARIABLE",
	}, nil
}

// GetItem fetches a single item by ID
func (c *CloverClient) GetItem(itemID string) (*Item, error) {
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	path := fmt.Sprintf("/v3/merchants/%s/items/%s?expand=itemStock", c.merchantID, itemID)
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
