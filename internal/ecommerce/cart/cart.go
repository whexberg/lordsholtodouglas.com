package cart

import "lsd3/internal/ecommerce"

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

// AddItem adds an item to the cart or increments quantity if exists.
// Quantity is capped at MaxItemQuantity and stock limit.
// If the item already exists but the price differs (variable-price re-add),
// the price is updated and quantity is set to 1 instead of incrementing.
func (c *Cart) AddItem(item CartItem) {
	// Variable-price items are always qty 1.
	if item.VariablePrice {
		item.Quantity = 1
	}
	cap := ecommerce.MaxQty(item.StockLimit)
	for i, existing := range c.Items {
		if existing.ProductID == item.ProductID {
			c.Items[i].StockLimit = item.StockLimit // refresh stock limit
			if item.VariablePrice {
				// Variable-price re-add: update price, keep qty 1.
				c.Items[i].PriceCents = item.PriceCents
				c.Items[i].Price = item.Price
				c.Items[i].Quantity = 1
				return
			}
			if existing.PriceCents != item.PriceCents {
				c.Items[i].PriceCents = item.PriceCents
				c.Items[i].Price = item.Price
				c.Items[i].Quantity = 1
				return
			}
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
			// Variable-price items are always qty 1.
			if item.VariablePrice {
				return
			}
			cap := ecommerce.MaxQty(item.StockLimit)
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
