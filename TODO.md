# LSD3 → Clamper Chapter Site: Implementation Todos

Transform from single-page Clover checkout into a multi-page ecommerce + content site for a Clamper (ECV) chapter. Products from Clover inventory, digital goods + donations only.

## Phase 1: Template Infrastructure
- [ ] Create `internal/templates/renderer.go` — template engine (load layout + partials + page, render by name)
- [ ] Create `templates/layout.html` — base layout with `{{block "content"}}`, head, nav, footer
- [ ] Create `templates/partials/nav.html` — nav bar (Home, History, Events, Members, Shop, Cart)
- [ ] Create `templates/partials/footer.html` — footer with chapter info
- [ ] Create `templates/home.html` — landing page with placeholder chapter name

## Phase 2: Content Pages + Updated Router
- [ ] Create `internal/handlers/pages.go` — PageHandler for Home, History, Events, Members
- [ ] Create `templates/history.html` — chapter & ECV history (placeholder content)
- [ ] Create `templates/events.html` — upcoming events page (placeholder content)
- [ ] Create `templates/members.html` — member info page (placeholder content)
- [ ] Update `cmd/server/main.go` — init renderer + page handler, add page routes, remove `/*` catch-all
- [ ] Update `static/style.css` — nav, footer, page layout, Clamper branding (earth tones)

## Phase 3: Shop
- [ ] Create `templates/shop.html` — product grid with "Add to Cart" forms
- [ ] Create `internal/handlers/shop.go` — ShopHandler calling Clover `ListItems()`, renders shop template
- [ ] Wire `GET /shop` route in main.go
- [ ] Wire existing `ProductHandler` API routes (`GET /api/products`, `GET /api/products/{id}`)

## Phase 4: Cart + Session
- [ ] Create `internal/middleware/session.go` — session cookie middleware (crypto/rand, no new deps)
- [ ] Create `internal/handlers/cart.go` — CartHandler using existing `CartStore` from `internal/models/cart.go`
- [ ] Create `templates/cart.html` — cart view with quantities, update/remove forms, total, checkout link
- [ ] Wire cart routes (`GET /cart`, `POST /cart/add`, `POST /cart/update`, `POST /cart/remove`) + session middleware

## Phase 5: Checkout Integration
- [ ] Create `templates/checkout.html` — migrate from `pages/index.html`, read cart total instead of hardcoded products
- [ ] Update `static/app.js` — conditionally init Clover SDK (only on checkout), get amount from cart
- [ ] Wire `GET /checkout` route

## Phase 6: Cleanup
- [ ] Remove `pages/index.html` (replaced by templates)
- [ ] Remove `internal/pages/` and `internal/templates/checkout.html` (unused prototypes)
- [ ] Update `deploy/Dockerfile` — COPY `templates/` instead of `pages/`
- [ ] Update `CLAUDE.md` to reflect new structure

---

## Key Context

**Existing code to reuse (already written, just needs wiring):**
- `internal/models/cart.go` — Cart, CartItem, CartStore (thread-safe, in-memory)
- `internal/models/product.go` — Product model + `ProductsFromCloverItems()` converter
- `internal/models/order.go` — Order models + Clover converters
- `internal/handlers/products.go` — ProductHandler (ListProducts, GetProduct)
- `internal/handlers/orders.go` — OrderHandler (CreateOrder, GetOrder, ListOrders)
- `internal/clover/clover.go` — ListItems(), GetItem(), ProcessPayment(), etc.

**Architecture decisions:**
- Go `html/template` for SSR (no frontend build tools needed)
- No new dependencies — `crypto/rand` for session IDs
- Cart total computed server-side at checkout (prevents price manipulation)
- Content pages are static templates (edit + redeploy)
- Products managed in Clover dashboard (no admin UI)
- Chapter name TBD — use placeholder for now
