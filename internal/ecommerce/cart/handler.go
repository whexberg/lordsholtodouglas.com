package cart

import (
	"encoding/json"
	"log"
	"lsd3/internal/middleware"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

// CartHandler handles cart actions and API.
type CartHandler struct {
	store  *SQLiteStore
	lookup ItemLookup
}

// NewCartHandler creates a new cart handler.
func NewCartHandler(store *SQLiteStore, lookup ItemLookup) *CartHandler {
	return &CartHandler{store: store, lookup: lookup}
}

type cartJSON struct {
	Items      []cartItemJSON `json:"items"`
	TotalCents int64          `json:"totalCents"`
	CartCount  int            `json:"cartCount"`
}

type cartItemJSON struct {
	ProductID     string  `json:"productId"`
	Name          string  `json:"name"`
	PriceCents    int64   `json:"priceCents"`
	Price         float64 `json:"price"`
	Quantity      int     `json:"quantity"`
	SubtotalCents int64   `json:"subtotalCents"`
	StockLimit    int     `json:"stockLimit"`    // -1 = unlimited
	VariablePrice bool    `json:"variablePrice"` // true for donation-style items
}

func buildCartJSON(cart *Cart) cartJSON {
	items := make([]cartItemJSON, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = cartItemJSON{
			ProductID:     item.ProductID,
			Name:          item.Name,
			PriceCents:    item.PriceCents,
			Price:         item.Price,
			Quantity:      item.Quantity,
			SubtotalCents: item.PriceCents * int64(item.Quantity),
			StockLimit:    item.StockLimit,
			VariablePrice: item.VariablePrice,
		}
	}
	return cartJSON{
		Items:      items,
		TotalCents: cart.Total(),
		CartCount:  cart.TotalItems(),
	}
}

func writeCartJSON(w http.ResponseWriter, cart *Cart) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildCartJSON(cart))
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(middleware.SessionIDKey).(string)
	cart := h.store.Get(sessionID)
	writeCartJSON(w, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(middleware.SessionIDKey).(string)

	productID := r.FormValue("product_id")
	if productID == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	// Fetch authoritative item data from inventory — never trust client values.
	info, err := h.lookup.LookupItem(productID)
	if err != nil {
		log.Printf("cart add: lookup item %s: %v", productID, err)
		http.Error(w, "Product not found", http.StatusBadRequest)
		return
	}

	priceCents := info.PriceCents
	if info.VariablePrice {
		customPrice := r.FormValue("custom_price")
		if customPrice == "" {
			http.Error(w, "Price is required for this item", http.StatusBadRequest)
			return
		}
		dollars, err := strconv.ParseFloat(customPrice, 64)
		if err != nil || dollars < 1.00 || dollars > 10000.00 {
			http.Error(w, "Price must be between $1.00 and $10,000.00", http.StatusBadRequest)
			return
		}
		priceCents = int64(math.Round(dollars * 100))
	}

	item := CartItem{
		ProductID:     productID,
		Name:          info.Name,
		PriceCents:    priceCents,
		Price:         float64(priceCents) / 100.0,
		Quantity:      1,
		StockLimit:    info.StockCount,
		VariablePrice: info.VariablePrice,
	}

	cart := h.store.Get(sessionID)
	cart.AddItem(item)
	if err := h.store.Save(sessionID, cart); err != nil {
		log.Printf("cart save error: %v", err)
		http.Error(w, "Failed to save cart", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		cart = h.store.Get(sessionID)
		writeCartJSON(w, cart)
		return
	}

	http.Redirect(w, r, safeRedirect(r), http.StatusSeeOther)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(middleware.SessionIDKey).(string)

	productID := r.FormValue("product_id")
	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	cart := h.store.Get(sessionID)
	cart.UpdateQuantity(productID, quantity)
	if err := h.store.Save(sessionID, cart); err != nil {
		log.Printf("cart save error: %v", err)
		http.Error(w, "Failed to save cart", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		cart = h.store.Get(sessionID)
		writeCartJSON(w, cart)
		return
	}

	http.Redirect(w, r, safeRedirect(r), http.StatusSeeOther)
}

func (h *CartHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(middleware.SessionIDKey).(string)

	productID := r.FormValue("product_id")
	if productID == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	customPrice := r.FormValue("custom_price")
	dollars, err := strconv.ParseFloat(customPrice, 64)
	if err != nil || dollars < 1.00 || dollars > 10000.00 {
		http.Error(w, "Price must be between $1.00 and $10,000.00", http.StatusBadRequest)
		return
	}
	priceCents := int64(math.Round(dollars * 100))

	cart := h.store.Get(sessionID)
	for i, item := range cart.Items {
		if item.ProductID == productID && item.VariablePrice {
			cart.Items[i].PriceCents = priceCents
			cart.Items[i].Price = float64(priceCents) / 100.0
			break
		}
	}
	if err := h.store.Save(sessionID, cart); err != nil {
		log.Printf("cart save error: %v", err)
		http.Error(w, "Failed to save cart", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		cart = h.store.Get(sessionID)
		writeCartJSON(w, cart)
		return
	}

	http.Redirect(w, r, safeRedirect(r), http.StatusSeeOther)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(middleware.SessionIDKey).(string)

	productID := r.FormValue("product_id")
	if productID == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	cart := h.store.Get(sessionID)
	cart.RemoveItem(productID)
	if err := h.store.Save(sessionID, cart); err != nil {
		log.Printf("cart save error: %v", err)
		http.Error(w, "Failed to save cart", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		cart = h.store.Get(sessionID)
		writeCartJSON(w, cart)
		return
	}

	http.Redirect(w, r, safeRedirect(r), http.StatusSeeOther)
}

// safeRedirect returns the Referer path if it's on the same host, otherwise "/shop".
func safeRedirect(r *http.Request) string {
	ref := r.Referer()
	if ref == "" {
		return "/shop"
	}
	u, err := url.Parse(ref)
	if err != nil || u.Host != "" && u.Host != r.Host {
		return "/shop"
	}
	return u.Path
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
