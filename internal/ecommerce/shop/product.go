package shop

import "lsd3/internal/ecommerce/clover"

// Product represents a storefront product mapped from Clover Item
type Product struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"` // in dollars for display
	PriceCents    int64   `json:"price_cents"`
	SKU           string  `json:"sku,omitempty"`
	Available     bool    `json:"available"`
	TrackStock    bool    `json:"track_stock"`
	StockCount    int     `json:"stock_count"`
	Description   string  `json:"description,omitempty"`
	VariablePrice bool    `json:"variable_price"`
}

// IsVariablePrice returns true if this product has user-set pricing.
func (p Product) IsVariablePrice() bool {
	return p.VariablePrice
}

// ProductFromCloverItem converts a Clover Item to a Product
func ProductFromCloverItem(item clover.Item) Product {
	return Product{
		ID:            item.ID,
		Name:          item.Name,
		Price:         float64(item.Price) / 100.0,
		PriceCents:    item.Price,
		SKU:           item.SKU,
		Available:     item.Available && !item.Hidden,
		TrackStock:    item.AutoManageStock,
		StockCount:    item.StockCount(),
		VariablePrice: item.PriceType == "VARIABLE",
	}
}

// LowStock returns true if stock is tracked and fewer than 25 remain.
func (p Product) LowStock() bool {
	return p.TrackStock && p.StockCount < 25
}

// ProductsFromCloverItems converts a slice of Clover Items to Products
func ProductsFromCloverItems(items []clover.Item) []Product {
	products := make([]Product, 0, len(items))
	for _, item := range items {
		if !item.Hidden {
			products = append(products, ProductFromCloverItem(item))
		}
	}
	return products
}
