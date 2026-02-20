package clover

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RefundRequest represents a Clover refund request
type RefundRequest struct {
	PaymentID string `json:"paymentId"`
	Amount    int64  `json:"amount,omitempty"` // partial refund amount in cents, omit for full refund
	Reason    string `json:"reason,omitempty"`
}

// RefundResponse represents a Clover refund response
type RefundResponse struct {
	ID        string `json:"id"`
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdTime,omitempty"`
}

// CreateRefund processes a refund through Clover
func (c *CloverClient) CreateRefund(req RefundRequest) (*RefundResponse, error) {
	if c.PrivateKey == "" || c.merchantID == "" {
		return nil, fmt.Errorf("missing Clover credentials")
	}

	if req.PaymentID == "" {
		return nil, fmt.Errorf("payment ID required")
	}

	path := fmt.Sprintf("/v3/merchants/%s/refunds", c.merchantID)
	body, status, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var refundResp RefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return &refundResp, fmt.Errorf("clover API error: %d", status)
	}

	return &refundResp, nil
}
