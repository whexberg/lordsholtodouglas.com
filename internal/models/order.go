package models

import "lsd3/internal/clover"

// OrderItem represents a line item in an order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// Order represents a storefront order mapped from Clover Order
type Order struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Total     float64     `json:"total"`
	Currency  string      `json:"currency"`
	Items     []OrderItem `json:"items"`
	CreatedAt int64       `json:"created_at,omitempty"`
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	Items []OrderItem `json:"items"`
	Note  string      `json:"note,omitempty"`
}

// OrderFromCloverOrder converts a Clover Order to a storefront Order
func OrderFromCloverOrder(co clover.Order) Order {
	items := make([]OrderItem, 0, len(co.LineItems))
	for _, li := range co.LineItems {
		items = append(items, OrderItem{
			ProductID: li.ItemID,
			Name:      li.Name,
			Price:     float64(li.Price) / 100.0,
			Quantity:  li.Quantity,
		})
	}

	return Order{
		ID:        co.ID,
		Status:    co.State,
		Total:     float64(co.Total) / 100.0,
		Currency:  co.Currency,
		Items:     items,
		CreatedAt: co.CreatedTime,
	}
}

// ToCloverCreateOrderRequest converts a CreateOrderRequest to a Clover request
func (r CreateOrderRequest) ToCloverCreateOrderRequest() clover.CreateOrderRequest {
	lineItems := make([]clover.OrderLineItem, 0, len(r.Items))
	for _, item := range r.Items {
		lineItems = append(lineItems, clover.OrderLineItem{
			ItemID:   item.ProductID,
			Name:     item.Name,
			Price:    int64(item.Price * 100),
			Quantity: item.Quantity,
		})
	}

	return clover.CreateOrderRequest{
		Currency:  "USD",
		Note:      r.Note,
		LineItems: lineItems,
	}
}
