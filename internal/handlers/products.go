package handlers

import (
	"encoding/json"
	"log"
	"lsd3/internal/clover"
	"lsd3/internal/models"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ProductHandler handles product-related requests
type ProductHandler struct {
	clover *clover.CloverClient
}

// NewProductHandler creates a new product handler
func NewProductHandler(c *clover.CloverClient) *ProductHandler {
	return &ProductHandler{clover: c}
}

// ListProducts handles GET /api/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.clover.ListItems()
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	products := models.ProductsFromCloverItems(items)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProduct handles GET /api/products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	if productID == "" {
		http.Error(w, "Product ID required", http.StatusBadRequest)
		return
	}

	item, err := h.clover.GetItem(productID)
	if err != nil {
		log.Printf("Error fetching product %s: %v", productID, err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	product := models.ProductFromCloverItem(*item)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
