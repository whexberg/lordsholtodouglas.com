package handlers

import (
	"encoding/json"
	"log"
	"lsd3/internal/clover"
	"net/http"
)

// CheckoutHandler handles checkout and payment requests
type CheckoutHandler struct {
	clover *clover.CloverClient
}

// NewCheckoutHandler creates a new checkout handler
func NewCheckoutHandler(c *clover.CloverClient) *CheckoutHandler {
	return &CheckoutHandler{clover: c}
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CheckoutRequest represents a checkout request
type CheckoutRequest struct {
	Amount   int          `json:"amount"`  // in cents
	Token    string       `json:"token"`   // Clover payment token
	Product  string       `json:"product"` // product identifier
	Customer CustomerInfo `json:"customer"`
}

// CheckoutResponse represents a checkout response
type CheckoutResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Amount    int    `json:"amount"`
}

// GetConfig handles GET /api/config - returns public configuration
func (h *CheckoutHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]string{
		"publicKey":  h.clover.PublicKey,
		"merchantId": h.clover.MerchantID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// ProcessCheckout handles POST /api/checkout
func (h *CheckoutHandler) ProcessCheckout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Checkout decode error: %v", err)
		jsonError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		jsonError(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	if req.Amount > 999999 { // Max $9,999.99
		jsonError(w, "Amount exceeds maximum", http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		jsonError(w, "Payment token required", http.StatusBadRequest)
		return
	}
	if req.Customer.Email == "" {
		jsonError(w, "Customer email required", http.StatusBadRequest)
		return
	}

	log.Printf("Processing payment: amount_cents=%d product=%s customer=%s", req.Amount, req.Product, req.Customer.Email)

	paymentReq := clover.ChargeRequest{
		Amount:   req.Amount,
		Currency: "usd",
		Source:   req.Token,
	}

	payment, err := h.clover.ProcessPayment(paymentReq)
	if err != nil {
		log.Printf("Payment processing error: %v", err)
		jsonError(w, "Payment processing failed", http.StatusInternalServerError)
		return
	}

	resp := CheckoutResponse{
		PaymentID: payment.ID,
		Status:    payment.Status,
		Amount:    payment.Amount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
