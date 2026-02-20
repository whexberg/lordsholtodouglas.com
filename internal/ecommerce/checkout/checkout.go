package checkout

import "lsd3/internal/ecommerce/clover"

// OrderService defines order management operations used by checkout.
type OrderService interface {
	CreateOrder(req clover.CreateOrderRequest) (*clover.Order, error)
	AddLineItemToOrder(orderID string, item clover.OrderLineItem) (*clover.OrderLineItem, error)
	DeleteOrder(orderID string) error
}

// PaymentProcessor handles paying for orders.
type PaymentProcessor interface {
	PayForOrder(orderID string, source string, receiptEmail string) (*clover.ChargeResponse, error)
}

// ItemFetcher retrieves inventory item details for checkout.
type ItemFetcher interface {
	GetItem(itemID string) (*clover.Item, error)
}

// ConfigProvider provides public Clover configuration for the frontend SDK.
type ConfigProvider interface {
	GetPublicKey() string
	GetMerchantID() string
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Note      string `json:"note"`
}

// CheckoutRequest represents a checkout request
type CheckoutRequest struct {
	Token    string       `json:"token"` // Clover payment token
	Customer CustomerInfo `json:"customer"`
}

// CheckoutResponse represents a checkout response
type CheckoutResponse struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id,omitempty"`
	Status    string `json:"status"`
	Amount    int    `json:"amount"`
}
