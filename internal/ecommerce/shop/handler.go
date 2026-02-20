package shop

import (
	"encoding/json"
	"log"
	"net/http"

	"lsd3/internal/middleware"
	"lsd3/internal/templates"

	"github.com/go-chi/chi/v5"
)

// ShopHandler serves the shop page.
type ShopHandler struct {
	items    ItemLister
	renderer templates.Renderer
}

// NewShopHandler creates a new shop handler.
func NewShopHandler(items ItemLister, r templates.Renderer) *ShopHandler {
	return &ShopHandler{items: items, renderer: r}
}

// PageData holds common data passed to every page template.
type PageData struct {
	CartCount int
}

func pageDataFromRequest(r *http.Request) PageData {
	count, _ := r.Context().Value(middleware.CartCountKey).(int)
	return PageData{CartCount: count}
}

type shopData struct {
	PageData
	Products []Product
}

func (h *ShopHandler) Shop(w http.ResponseWriter, r *http.Request) {
	items, err := h.items.ListItems()
	if err != nil {
		log.Printf("shop: fetch items: %v", err)
		items = nil
	}

	products := ProductsFromCloverItems(items)

	data := shopData{
		PageData: pageDataFromRequest(r),
		Products: products,
	}

	if err := h.renderer.Render(w, "shop.html", data); err != nil {
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
