package cart

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID     string  `json:"product_id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	PriceCents    int64   `json:"price_cents"`
	Quantity      int     `json:"quantity"`
	StockLimit    int     `json:"stock_limit"`    // -1 = unlimited (not tracked)
	VariablePrice bool    `json:"variable_price"` // true for donation-style items
}

// Subtotal returns the line item subtotal in dollars.
func (ci *CartItem) Subtotal() float64 {
	return float64(ci.PriceCents*int64(ci.Quantity)) / 100.0
}

// ItemInfo holds authoritative item data from inventory.
type ItemInfo struct {
	Name          string
	PriceCents    int64
	StockCount    int  // -1 means unlimited (stock not tracked)
	VariablePrice bool // true if user sets the price (e.g. donations)
}

// ItemLookup fetches an item's authoritative data by ID.
type ItemLookup interface {
	LookupItem(itemID string) (*ItemInfo, error)
}
