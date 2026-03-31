package shop

import (
	"encoding/json"
	"log"
	"net/http"

	"lsd3/internal/data_view"
	"lsd3/internal/templates"

	"github.com/go-chi/chi/v5"
)

// ShopHandler serves the shop page.
type ShopHandler struct {
	items ItemLister
}

// NewShopHandler creates a new shop handler.
func NewShopHandler(items ItemLister) *ShopHandler {
	return &ShopHandler{items: items}
}

func (h *ShopHandler) Shop(w http.ResponseWriter, r *http.Request) {
	items, err := h.items.ListItems()
	if err != nil {
		log.Printf("shop: fetch items: %v", err)
		items = nil
	}

	products := ProductsFromCloverItems(items)

	views := make([]data_view.ProductView, len(products))
	for i, p := range products {
		views[i] = data_view.ProductView{
			ID:            p.ID,
			Name:          p.Name,
			Price:         p.Price,
			PriceCents:    p.PriceCents,
			Available:     p.Available,
			TrackStock:    p.TrackStock,
			StockCount:    p.StockCount,
			Description:   p.Description,
			VariablePrice: p.VariablePrice,
		}
	}

	data := data_view.ShopData{
		PageData: data_view.PageDataFromRequest(r),
		Products: views,
	}

	if err := templates.ShopPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render shop: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ProductHandler handles product API requests.
type ProductHandler struct {
	items ItemLister
}

// NewProductHandler creates a new product handler.
func NewProductHandler(items ItemLister) *ProductHandler {
	return &ProductHandler{items: items}
}

// ListProducts handles GET /api/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.items.ListItems()
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	products := ProductsFromCloverItems(items)

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

	item, err := h.items.GetItem(productID)
	if err != nil {
		log.Printf("Error fetching product %s: %v", productID, err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	product := ProductFromCloverItem(*item)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
