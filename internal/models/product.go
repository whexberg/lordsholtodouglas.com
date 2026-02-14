package models

import "lsd3/internal/clover"

// Product represents a storefront product mapped from Clover Item
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"` // in dollars for display
	PriceCents  int64   `json:"price_cents"`
	SKU         string  `json:"sku,omitempty"`
	Available   bool    `json:"available"`
	StockCount  int     `json:"stock_count"`
	Description string  `json:"description,omitempty"`
}

// ProductFromCloverItem converts a Clover Item to a Product
func ProductFromCloverItem(item clover.Item) Product {
	return Product{
		ID:         item.ID,
		Name:       item.Name,
		Price:      float64(item.Price) / 100.0,
		PriceCents: item.Price,
		SKU:        item.SKU,
		Available:  item.Available && !item.Hidden,
		StockCount: item.StockCount,
	}
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
