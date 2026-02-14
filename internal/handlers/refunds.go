package handlers

import (
	"encoding/json"
	"log"
	"lsd3/internal/clover"
	"net/http"
)

// RefundHandler handles refund requests
type RefundHandler struct {
	clover *clover.CloverClient
}

// NewRefundHandler creates a new refund handler
func NewRefundHandler(c *clover.CloverClient) *RefundHandler {
	return &RefundHandler{clover: c}
}

// RefundRequest represents a refund request
type RefundRequest struct {
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount,omitempty"` // in cents, omit for full refund
	Reason    string `json:"reason,omitempty"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	RefundID  string `json:"refund_id"`
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
}

// ProcessRefund handles POST /api/refunds
func (h *RefundHandler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	var req RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Refund decode error: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.PaymentID == "" {
		http.Error(w, "Payment ID required", http.StatusBadRequest)
		return
	}

	if req.Amount < 0 {
		http.Error(w, "Invalid refund amount", http.StatusBadRequest)
		return
	}

	log.Printf("Processing refund: payment_id=%s amount=%d", req.PaymentID, req.Amount)

	cloverReq := clover.RefundRequest{
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Reason:    req.Reason,
	}

	refund, err := h.clover.CreateRefund(cloverReq)
	if err != nil {
		log.Printf("Refund processing error: %v", err)
		http.Error(w, "Refund processing failed", http.StatusInternalServerError)
		return
	}

	resp := RefundResponse{
		RefundID:  refund.ID,
		PaymentID: refund.PaymentID,
		Amount:    refund.Amount,
		Status:    refund.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
