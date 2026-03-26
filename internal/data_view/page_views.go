package data_view

import (
	"html/template"

	"lsd3/internal/content"
	"lsd3/internal/ecommerce/cart"
)

type EventDetailData struct {
	PageData
	Event *content.Event
}

type HumbuggeryData struct {
	PageData
	Intro   string
	Humbugs []content.Humbug
}

type HumbugDetailData struct {
	PageData
	Humbug *content.Humbug
}

type HistoryReportsData struct {
	PageData
	Reports []content.HistoryReport
}

type HistoryReportData struct {
	PageData
	Report *content.HistoryReport
}

type StaticPageData struct {
	PageData
	Description string
	ContentHTML template.HTML
}

// ProductView is a template-friendly projection of a shop product.
type ProductView struct {
	ID            string
	Name          string
	Price         float64
	PriceCents    int64
	Available     bool
	TrackStock    bool
	StockCount    int
	Description   string
	VariablePrice bool
}

// IsVariablePrice returns true if this product has user-set pricing.
func (p ProductView) IsVariablePrice() bool {
	return p.VariablePrice
}

// LowStock returns true if stock is tracked and fewer than 25 remain.
func (p ProductView) LowStock() bool {
	return p.TrackStock && p.StockCount < 25
}

type ShopData struct {
	PageData
	Products []ProductView
}

type CheckoutData struct {
	PageData
	Cart               *cart.Cart
	CloverSDKURL       string
	ProcessingFeeCents int
	GrandTotalCents    int64
}
