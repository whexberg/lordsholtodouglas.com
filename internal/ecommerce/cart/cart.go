package cart

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID  string  `json:"product_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	PriceCents int64   `json:"price_cents"`
	Quantity   int     `json:"quantity"`
	StockLimit int     `json:"stock_limit"` // -1 = unlimited (not tracked)
}

// Subtotal returns the line item subtotal in dollars.
func (ci *CartItem) Subtotal() float64 {
	return float64(ci.PriceCents*int64(ci.Quantity)) / 100.0
}

// Cart represents a shopping cart
type Cart struct {
	ID    string     `json:"id"`
	Items []CartItem `json:"items"`
}

// Total calculates the cart total in cents
func (c *Cart) Total() int64 {
	var total int64
	for _, item := range c.Items {
		total += item.PriceCents * int64(item.Quantity)
	}
	return total
}

// TotalDollars returns the cart total as a float for display
func (c *Cart) TotalDollars() float64 {
	return float64(c.Total()) / 100.0
}

// MaxItemQuantity is the maximum quantity allowed per cart item.
const MaxItemQuantity = 99

// maxQty returns the effective maximum quantity for an item,
// considering both MaxItemQuantity and stock limits.
func maxQty(stockLimit int) int {
	if stockLimit >= 0 && stockLimit < MaxItemQuantity {
		return stockLimit
	}
	return MaxItemQuantity
}

// AddItem adds an item to the cart or increments quantity if exists.
// Quantity is capped at MaxItemQuantity and stock limit.
func (c *Cart) AddItem(item CartItem) {
	cap := maxQty(item.StockLimit)
	for i, existing := range c.Items {
		if existing.ProductID == item.ProductID {
			c.Items[i].StockLimit = item.StockLimit // refresh stock limit
			c.Items[i].Quantity += item.Quantity
			if c.Items[i].Quantity > cap {
				c.Items[i].Quantity = cap
			}
			return
		}
	}
	if item.Quantity > cap {
		item.Quantity = cap
	}
	c.Items = append(c.Items, item)
}

// RemoveItem removes an item from the cart
func (c *Cart) RemoveItem(productID string) {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return
		}
	}
}

// UpdateQuantity updates the quantity of an item in the cart.
// Quantity is capped at MaxItemQuantity and stock limit; values ≤ 0 remove the item.
func (c *Cart) UpdateQuantity(productID string, quantity int) {
	if quantity <= 0 {
		c.RemoveItem(productID)
		return
	}
	for i, item := range c.Items {
		if item.ProductID == productID {
			cap := maxQty(item.StockLimit)
			if quantity > cap {
				quantity = cap
			}
			c.Items[i].Quantity = quantity
			return
		}
	}
}

// TotalItems returns the sum of all item quantities in the cart.
func (c *Cart) TotalItems() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

// Clear empties the cart
func (c *Cart) Clear() {
	c.Items = []CartItem{}
}

// ItemInfo holds authoritative item data from inventory.
type ItemInfo struct {
	Name       string
	PriceCents int64
	StockCount int  // -1 means unlimited (stock not tracked)
}

// ItemLookup fetches an item's authoritative data by ID.
type ItemLookup interface {
	LookupItem(itemID string) (*ItemInfo, error)
}
