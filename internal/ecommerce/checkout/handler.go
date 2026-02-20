package checkout

import (
	"encoding/json"
	"fmt"
	"log"
	"lsd3/internal/ecommerce/cart"
	"lsd3/internal/ecommerce/clover"
	"lsd3/internal/middleware"
	"lsd3/internal/templates"
	"net/http"
	"strings"
)

// CheckoutHandler handles checkout and payment requests
type CheckoutHandler struct {
	orders OrderService
	pay    PaymentProcessor
	config ConfigProvider
	items  ItemFetcher
	carts  *cart.SQLiteStore
}

// NewCheckoutHandler creates a new checkout handler
func NewCheckoutHandler(
	orders OrderService,
	pay PaymentProcessor,
	config ConfigProvider,
	items ItemFetcher,
	carts *cart.SQLiteStore,
) *CheckoutHandler {
	return &CheckoutHandler{
		orders: orders,
		pay:    pay,
		config: config,
		items:  items,
		carts:  carts,
	}
}

// GetConfig handles GET /api/config - returns public configuration
func (h *CheckoutHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]string{
		"publicKey":  h.config.GetPublicKey(),
		"merchantId": h.config.GetMerchantID(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

const maxCustomerFieldLen = 200
const maxNoteLen = 500

// ProcessingFeeCents is the flat processing fee added to every order.
const ProcessingFeeCents = 100

// ProcessCheckout handles POST /api/checkout
func (h *CheckoutHandler) ProcessCheckout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Checkout decode error: %v", err)
		jsonError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		jsonError(w, "Payment token required", http.StatusBadRequest)
		return
	}

	if req.Customer.FirstName == "" || req.Customer.LastName == "" {
		jsonError(w, "Customer name required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Customer.Email) {
		jsonError(w, "Valid customer email required", http.StatusBadRequest)
		return
	}

	if len(req.Customer.FirstName) > maxCustomerFieldLen ||
		len(req.Customer.LastName) > maxCustomerFieldLen ||
		len(req.Customer.Email) > maxCustomerFieldLen ||
		len(req.Customer.Phone) > maxCustomerFieldLen {
		jsonError(w, "Customer field too long", http.StatusBadRequest)
		return
	}
	if len(req.Customer.Note) > maxNoteLen {
		jsonError(w, "Note too long (max 500 characters)", http.StatusBadRequest)
		return
	}

	// Sanitize customer inputs — strip control chars, collapse whitespace.
	req.Customer.FirstName = sanitize(req.Customer.FirstName)
	req.Customer.LastName = sanitize(req.Customer.LastName)
	req.Customer.Phone = sanitize(req.Customer.Phone)
	req.Customer.Note = sanitize(req.Customer.Note)

	// Read cart server-side — ignore client-sent amount
	sessionID, _ := r.Context().Value(middleware.SessionIDKey).(string)
	if sessionID == "" {
		jsonError(w, "No session", http.StatusBadRequest)
		return
	}

	c := h.carts.Get(sessionID)
	if len(c.Items) == 0 {
		jsonError(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	cartTotal := int(c.Total())
	if cartTotal <= 0 || cartTotal > 999999 {
		jsonError(w, "Invalid cart total", http.StatusBadRequest)
		return
	}
	amount := cartTotal + ProcessingFeeCents

	log.Printf("Processing checkout: amount_cents=%d (cart=%d + fee=%d) items=%d", amount, cartTotal, ProcessingFeeCents, len(c.Items))

	// Create Clover order with line items referencing inventory.
	orderReq := clover.CreateOrderRequest{
		Currency: "USD",
		Note:     buildOrderNote(req.Customer),
	}
	o, err := h.orders.CreateOrder(orderReq)
	if err != nil {
		log.Printf("Order creation error: %v", err)
		jsonError(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	for _, item := range c.Items {
		// Fetch item to determine pricing type and check stock.
		cloverItem, err := h.items.GetItem(item.ProductID)
		if err != nil {
			log.Printf("Fetch item error (order %s, item %s): %v", o.ID, item.ProductID, err)
			h.cleanupOrder(o.ID)
			jsonError(w, "Failed to look up item", http.StatusInternalServerError)
			return
		}

		// Reject if stock is tracked and insufficient.
		if cloverItem.AutoManageStock && item.Quantity > cloverItem.StockCount() {
			log.Printf("Insufficient stock for %s (%s): requested %d, available %d",
				cloverItem.Name, item.ProductID, item.Quantity, cloverItem.StockCount())
			h.cleanupOrder(o.ID)
			msg := fmt.Sprintf("Not enough stock for %s (only %d available)", cloverItem.Name, cloverItem.StockCount())
			jsonError(w, msg, http.StatusConflict)
			return
		}

		if cloverItem.PriceType == "PER_UNIT" {
			// Per-unit items support unitQty in a single call.
			lineItem := clover.OrderLineItem{
				Item:    &clover.ItemRef{ID: item.ProductID},
				UnitQty: item.Quantity * 1000, // Clover uses thousandths (1000 = 1 unit)
			}
			if _, err := h.orders.AddLineItemToOrder(o.ID, lineItem); err != nil {
				log.Printf("Add line item error (order %s, item %s): %v", o.ID, item.ProductID, err)
				h.cleanupOrder(o.ID)
				jsonError(w, "Failed to add items to order", http.StatusInternalServerError)
				return
			}
		} else {
			// Fixed-price items require one API call per unit.
			lineItem := clover.OrderLineItem{
				Item: &clover.ItemRef{ID: item.ProductID},
			}
			for q := 0; q < item.Quantity; q++ {
				if _, err := h.orders.AddLineItemToOrder(o.ID, lineItem); err != nil {
					log.Printf("Add line item error (order %s, item %s): %v", o.ID, item.ProductID, err)
					h.cleanupOrder(o.ID)
					jsonError(w, "Failed to add items to order", http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// Add processing fee as a non-inventory line item.
	feeItem := clover.OrderLineItem{
		Name:    "Processing Fee",
		Price:   ProcessingFeeCents,
		UnitQty: 1000, // qty=1 in Clover thousandths
	}
	if _, err := h.orders.AddLineItemToOrder(o.ID, feeItem); err != nil {
		log.Printf("Add processing fee error (order %s): %v", o.ID, err)
		h.cleanupOrder(o.ID)
		jsonError(w, "Failed to add processing fee to order", http.StatusInternalServerError)
		return
	}

	// Pay for the order — links payment to line items for itemized receipts.
	payment, err := h.pay.PayForOrder(o.ID, req.Token, req.Customer.Email)
	if err != nil {
		log.Printf("Payment processing error: %v", err)
		h.cleanupOrder(o.ID)
		jsonError(w, "Payment processing failed", http.StatusInternalServerError)
		return
	}

	// Clear cart on success
	h.carts.Delete(sessionID)

	resp := CheckoutResponse{
		PaymentID: payment.ID,
		OrderID:   o.ID,
		Status:    payment.Status,
		Amount:    payment.Amount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// cleanupOrder attempts to delete a Clover order that failed during checkout.
func (h *CheckoutHandler) cleanupOrder(orderID string) {
	if err := h.orders.DeleteOrder(orderID); err != nil {
		log.Printf("Failed to clean up order %s: %v", orderID, err)
	}
}

// isValidEmail performs basic email format validation.
func isValidEmail(email string) bool {
	if len(email) == 0 || len(email) > maxCustomerFieldLen {
		return false
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	return len(local) > 0 && len(domain) > 2 && strings.Contains(domain, ".")
}

// sanitize strips control characters and collapses whitespace.
func sanitize(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r < 0x20 && r != '\n' { // allow newlines for notes
			continue
		}
		if r == ' ' || r == '\t' {
			if !prevSpace {
				b.WriteByte(' ')
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// buildOrderNote formats customer info for the Clover order note (visible on receipts).
func buildOrderNote(c CustomerInfo) string {
	note := c.FirstName + " " + c.LastName + "\n" + c.Email
	if c.Phone != "" {
		note += "\n" + c.Phone
	}
	if c.Note != "" {
		note += "\n\n" + c.Note
	}
	return note
}

// CheckoutPageHandler serves the checkout page with cart data.
type CheckoutPageHandler struct {
	store     *cart.SQLiteStore
	renderer  templates.Renderer
	cloverSDK string
}

// NewCheckoutPageHandler creates a new checkout page handler.
func NewCheckoutPageHandler(store *cart.SQLiteStore, r templates.Renderer, cloverSDKURL string) *CheckoutPageHandler {
	return &CheckoutPageHandler{store: store, renderer: r, cloverSDK: cloverSDKURL}
}

// PageData holds common data passed to every page template.
type PageData struct {
	CartCount int
}

type checkoutPageData struct {
	PageData
	Cart              *cart.Cart
	CloverSDKURL      string
	ProcessingFeeCents int
	GrandTotalCents   int64
}

func (h *CheckoutPageHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(middleware.SessionIDKey).(string)

	count, _ := r.Context().Value(middleware.CartCountKey).(int)

	c := h.store.Get(sessionID)
	data := checkoutPageData{
		PageData:           PageData{CartCount: count},
		Cart:               c,
		CloverSDKURL:       h.cloverSDK,
		ProcessingFeeCents: ProcessingFeeCents,
		GrandTotalCents:    c.Total() + int64(ProcessingFeeCents),
	}

	if err := h.renderer.Render(w, "checkout.html", data); err != nil {
		log.Printf("render checkout: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
