package ecommerce

// MaxItemQuantity is the maximum quantity allowed per cart item.
const MaxItemQuantity = 99

// maxQty returns the effective maximum quantity for an item,
// considering both MaxItemQuantity and stock limits.
func MaxQty(stockLimit int) int {
	if stockLimit >= 0 && stockLimit < MaxItemQuantity {
		return stockLimit
	}
	return MaxItemQuantity
}
