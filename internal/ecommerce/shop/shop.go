package shop

import "lsd3/internal/ecommerce/clover"

// ItemLister fetches inventory items.
type ItemLister interface {
	ListItems() ([]clover.Item, error)
	GetItem(itemID string) (*clover.Item, error)
}
