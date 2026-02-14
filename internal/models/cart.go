package models

import "sync"

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID  string  `json:"product_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	PriceCents int64   `json:"price_cents"`
	Quantity   int     `json:"quantity"`
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

// AddItem adds an item to the cart or increments quantity if exists
func (c *Cart) AddItem(item CartItem) {
	for i, existing := range c.Items {
		if existing.ProductID == item.ProductID {
			c.Items[i].Quantity += item.Quantity
			return
		}
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

// UpdateQuantity updates the quantity of an item in the cart
func (c *Cart) UpdateQuantity(productID string, quantity int) {
	if quantity <= 0 {
		c.RemoveItem(productID)
		return
	}
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items[i].Quantity = quantity
			return
		}
	}
}

// Clear empties the cart
func (c *Cart) Clear() {
	c.Items = []CartItem{}
}

// CartStore provides thread-safe in-memory cart storage
type CartStore struct {
	mu    sync.RWMutex
	carts map[string]*Cart
}

// NewCartStore creates a new cart store
func NewCartStore() *CartStore {
	return &CartStore{
		carts: make(map[string]*Cart),
	}
}

// Get retrieves a cart by session ID, creating if not exists
func (s *CartStore) Get(sessionID string) *Cart {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cart, ok := s.carts[sessionID]; ok {
		return cart
	}

	cart := &Cart{
		ID:    sessionID,
		Items: []CartItem{},
	}
	s.carts[sessionID] = cart
	return cart
}

// Delete removes a cart by session ID
func (s *CartStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.carts, sessionID)
}
