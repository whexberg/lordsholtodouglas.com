package main

import (
	"log"
	"net/http"
	"time"

	"lsd3/internal/config"
	"lsd3/internal/content"
	"lsd3/internal/ecommerce/cart"
	"lsd3/internal/ecommerce/checkout"
	"lsd3/internal/ecommerce/clover"
	"lsd3/internal/ecommerce/shop"
	"lsd3/internal/middleware"
	"lsd3/internal/pages"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Load content from markdown files
	if err = content.Load("content"); err != nil {
		log.Fatalf("content: %v", err)
	}

	// Clover client
	cloverClient := clover.NewCloverClient(
		cfg.CloverBaseURL(),
		cfg.CloverEcommerceURL(),
		cfg.CloverMerchantID(),
		cfg.CloverPrivateKey(),
		cfg.CloverPublicKey(),
	)

	// Shared stores
	cartStore, err := cart.NewSQLiteStore("data/carts.db")
	if err != nil {
		log.Fatalf("cart store: %v", err)
	}
	defer cartStore.Close()

	// Handlers
	pageHandler := pages.NewPageHandler()
	shopHandler := shop.NewShopHandler(cloverClient)
	cartHandler := cart.NewCartHandler(cartStore, cloverClient)
	checkoutPageHandler := checkout.NewCheckoutPageHandler(cartStore, cfg.CloverSDKURL())
	checkoutHandler := checkout.NewCheckoutHandler(cloverClient, cloverClient, cloverClient, cloverClient, cartStore)
	productHandler := shop.NewProductHandler(cloverClient)

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RedirectSlashes)
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(middleware.Session(cartStore))
	r.Use(middleware.CSRF)
	r.Use(middleware.MaxBytes(1 << 20))                     // 1 MB
	rateLimiter, rateLimitMW := middleware.RateLimit(2, 10) // 2 req/s per IP, burst of 10
	defer rateLimiter.Close()
	r.Use(rateLimitMW)
	r.Use(middleware.SecureHeaders)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	})

	// 404 handler
	r.NotFound(pageHandler.NotFound)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Handle("/images/*", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images"))))

	// Content pages
	r.Get("/", pageHandler.Home)
	r.Get("/board-members", pageHandler.BoardMembers)
	r.Get("/board-members/{slug}", pageHandler.BoardMember)
	r.Get("/events", pageHandler.Events)
	r.Get("/events/{slug}", pageHandler.Event)
	r.Get("/events/{slug}/{date}", pageHandler.Event)
	r.Get("/humbuggery", pageHandler.Humbuggery)
	r.Get("/humbuggery/{slug}", pageHandler.Humbug)
	r.Get("/history-reports", pageHandler.HistoryReports)
	r.Get("/history-reports/{slug}", pageHandler.HistoryReport)

	// Shop
	r.Get("/shop", shopHandler.Shop)

	// Cart
	r.Post("/cart/add", cartHandler.AddItem)
	r.Post("/cart/update", cartHandler.UpdateItem)
	r.Post("/cart/update-price", cartHandler.UpdatePrice)
	r.Post("/cart/remove", cartHandler.RemoveItem)

	// Checkout
	r.Get("/checkout", checkoutPageHandler.Checkout)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/cart", cartHandler.GetCart)
		r.Post("/checkout", checkoutHandler.ProcessCheckout)
		r.Get("/config", checkoutHandler.GetConfig)
		r.Get("/products", productHandler.ListProducts)
		r.Get("/products/{id}", productHandler.GetProduct)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port(),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Server running on http://localhost:%s", cfg.Port())
	log.Fatal(server.ListenAndServe())
}
