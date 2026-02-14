package handlers

import (
	"encoding/json"
	"log"
	"lsd3/internal/clover"
	"lsd3/internal/models"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// OrderHandler handles order-related requests
type OrderHandler struct {
	clover *clover.CloverClient
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(c *clover.CloverClient) *OrderHandler {
	return &OrderHandler{clover: c}
}

// CreateOrder handles POST /api/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Order decode error: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "Order must have at least one item", http.StatusBadRequest)
		return
	}

	cloverReq := req.ToCloverCreateOrderRequest()
	cloverOrder, err := h.clover.CreateOrder(cloverReq)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	order := models.OrderFromCloverOrder(*cloverOrder)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetOrder handles GET /api/orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	cloverOrder, err := h.clover.GetOrder(orderID)
	if err != nil {
		log.Printf("Error fetching order %s: %v", orderID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	order := models.OrderFromCloverOrder(*cloverOrder)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ListOrders handles GET /api/orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	cloverOrders, err := h.clover.ListOrders()
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	orders := make([]models.Order, 0, len(cloverOrders))
	for _, co := range cloverOrders {
		orders = append(orders, models.OrderFromCloverOrder(co))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
