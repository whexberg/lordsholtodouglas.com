# LSD3 Project

Go web application — Lord Sholto Douglas Chapter #3 (E Clampus Vitus) ecommerce and content site powered by Clover.

## Structure

```
cmd/server/main.go               - Entry point, router setup
internal/clover/                  - Concrete Clover API client (unchanged)
  client.go                       - CloverClient struct, constructor, doRequest helper
  inventory.go                    - Item types, ListItems, GetItem
  orders.go                       - Order types, CreateOrder, GetOrder, ListOrders, AddLineItemToOrder
  payments.go                     - ChargeResponse, PayForOrder (Ecommerce API)
  refunds.go                      - RefundRequest/Response, CreateRefund
content/                              - Markdown content files (parsed at startup)
  _index.md                           - Site config (title, subtitle)
  board-members/_index.md             - Board members list in frontmatter
  sponsors/*.md                       - One file per sponsor
  events/*.md                         - Events with schedule frontmatter + markdown body
  history-reports/*.md                 - Reports with markdown body
  humbuggery/_index.md                - Image list in frontmatter
internal/content/                 - Content types and markdown loader
  content.go                      - Type definitions + exported package vars
  loader.go                       - Parse markdown at startup → populate package vars
  schedule.go                     - Recurrence logic (monthly/yearly schedule rules)
internal/templates/               - HTML renderer + Renderer interface
  renderer.go                     - Renderer interface + HTMLRenderer implementation
internal/pages/                   - Content pages feature (home, board, events, humbuggery, history)
  handler.go                      - PageHandler
internal/shop/                    - Product browsing feature
  handler.go                      - ShopHandler + ProductHandler
  product.go                      - Product model + Clover conversion
  shop.go                         - ItemLister interface
internal/middleware/               - HTTP middleware
  session.go                      - Session cookie + cart count injection
  csrf.go                         - Origin/Referer CSRF protection
internal/cart/                    - Cart models, SQLite store, handlers
  cart.go                         - Cart/CartItem models
  handler.go                      - CartHandler (add/update/remove/get)
  store.go                        - SQLite persistence
internal/checkout/                - Payment processing, orders, checkout page
  checkout.go                     - Interfaces (OrderService, PaymentProcessor, ConfigProvider) + request/response types
  handler.go                      - CheckoutHandler + CheckoutPageHandler
templates/                        - HTML templates (layout + partials + pages)
static/                           - JS/CSS assets served at /static/*
static/images/                    - Site images (gitignored, deploy separately)
docs/
  SECURITY.md                     - Security guide for hosting
deploy/
  Dockerfile                      - Production multi-stage build
  Dockerfile.dev                  - Dev image (air hot reload)
  docker-compose.yml              - Dev environment (hot reload + Caddy)
  docker-compose.prod.yml         - Production environment (compiled + Caddy + TLS)
  Caddyfile.dev                   - Caddy config for local dev
  Caddyfile.prod                  - Caddy config for production (auto TLS)
```

## Running

```bash
# Required environment variables (or use .env file)
export CLOVER_BASE_URL=https://sandbox.dev.clover.com  # or https://api.clover.com
export CLOVER_ECOMMERCE_URL=https://scl-sandbox.dev.clover.com  # or https://scl.clover.com
export CLOVER_MERCHANT_ID=...
export CLOVER_PUBLIC_API_TOKEN=...   # For frontend SDK
export CLOVER_PRIVATE_API_TOKEN=...  # For backend API
export PORT=8080                     # optional

go run ./cmd/server
```

## Routes

### Pages
- `GET /` — Home page (title, logo, Clamper's Creed, sponsors grid)
- `GET /board-members` — Board members with photos (3-col grid)
- `GET /events` — Upcoming events (server-side rendered list)
- `GET /humbuggery` — Hall of Humbuggery photo gallery
- `GET /history-reports` — History reports (2-col card grid)
- `GET /history-reports/{slug}` — Individual history report article
- `GET /shop` — Product grid (from Clover inventory)
- `GET /cart` — Shopping cart
- `GET /checkout` — Payment form with cart summary

### Navigation
Home | The Board | Events | Hall of Humbuggery | History Reports | Shop | Cart(N)

### Cart Actions
- `POST /cart/add` — Add item to cart
- `POST /cart/update` — Update item quantity
- `POST /cart/remove` — Remove item from cart

### API
- `GET /api/config` — Clover public key for frontend SDK
- `POST /api/checkout` — Process payment with tokenized card
- `POST /api/refunds` — Process refund
- `GET /api/products` — JSON product list (from Clover inventory)
- `GET /api/products/{id}` — JSON product detail

## Architecture

- **Server-side rendering** via Go `html/template` with shared layout
- **Content from markdown** in `content/` — parsed once at startup with goldmark (same renderer as Hugo). Board members, sponsors, events, history reports, humbuggery are YAML frontmatter + markdown body
- **Event schedules** support `once`, `monthly`, and `yearly` recurrence with multi-schedule events (e.g. time change mid-year). Display dates computed at load time
- **Products from Clover** inventory API — managed in Clover dashboard
- **SQLite cart** with session cookies (stdlib `crypto/rand`)
- **Cart total computed server-side** at checkout to prevent price manipulation
- **Stock managed by Clover** — auto-decremented when orders with inventory items are paid
- **Dark mode** with `.dark` class on `<html>`, persisted in localStorage, FOUC-free via `theme.js`
- **Minimal dependencies** — chi router, godotenv, goldmark (markdown), yaml.v3

## Content Management

Site content lives in `content/` as markdown files with YAML frontmatter (same format
as the Hugo site at `hugo-site/content/`). Content is parsed once at startup by
`internal/content/loader.go` using goldmark (the same markdown renderer Hugo uses).

To update content, edit the relevant `.md` file in `content/` and restart the server.
No Go code changes needed for content updates.

- **Board members**: Edit `content/board-members/_index.md` frontmatter `members` array
- **Sponsors**: Add/edit files in `content/sponsors/` (one `.md` per sponsor)
- **Events**: Add/edit files in `content/events/` with `schedules` frontmatter for dates
- **History reports**: Add/edit files in `content/history-reports/` (frontmatter + markdown body)
- **Humbuggery**: Edit `content/humbuggery/_index.md` frontmatter `images` array

## How It All Works

### Product Management

Products are managed entirely in the Clover merchant dashboard. The app fetches them
via `GET /v3/merchants/{mId}/items?expand=itemStock`. The `Product` model is a
projection of Clover's `Item` — hidden items are filtered out. Stock counts come from
the nested `itemStock.quantity` field (Clover returns this as a float).

### Session & Cart

Every visitor gets a session cookie (`session_id`, 32 hex chars from `crypto/rand`).
The session middleware (`internal/middleware/session.go`) sets the cookie and injects
the session ID + cart item count into the request context.

Carts live in an `SQLiteStore` (SQLite database in `data/carts.db`). Cart data is
keyed by session ID. Items in the cart store the product ID, name, unit price in cents,
and quantity. Carts are persisted to SQLite and survive server restarts.

### Shopping Flow

1. **Browse** — `GET /shop` calls `CloverClient.ListItems()`, converts to `Product`
   models, renders `shop.html` with a product grid.
2. **Add to cart** — Each product card has a form that POSTs to `/cart/add` with the
   product ID, name, and price in cents. The handler adds the item to the session's
   cart (or increments quantity if it already exists) and redirects to `/cart`.
3. **View/edit cart** — `GET /cart` renders the cart with quantities, subtotals, and
   update/remove forms. All mutations redirect back to `/cart`.

### Checkout Flow

**Frontend side** (`templates/checkout.html` + `static/app.js`):

1. Page renders cart summary and a hidden input with the cart total in cents.
2. JavaScript loads the Clover SDK with the public key from `GET /api/config`.
3. Clover's iframe captures card details (number, date, CVV, postal code).
4. On submit, `clover.createToken()` tokenizes the card client-side.
5. JavaScript POSTs the token + customer info to `POST /api/checkout`.

**Backend side** (`internal/checkout/handler.go`):

1. **Read cart server-side** — ignores any client-sent amount. The cart total is
   computed from the SQLiteStore to prevent price manipulation.
2. **Create Clover order** — `POST /v3/merchants/{mId}/orders` with customer email
   as the note.
3. **Add line items** — for each cart item, `POST /v3/merchants/{mId}/orders/{orderId}/line_items`
   with `item.id` referencing the inventory item. Clover pulls the name and price
   from inventory. `unitQty` is quantity * 1000 (Clover uses thousandths).
4. **Pay for order** — `POST /v1/orders/{orderId}/pay` on the Ecommerce API with
   the card token. This links the payment to the order so Clover receipts show
   itemized line items (not generic "Item 1").
5. **Clear cart** — deletes the session's cart from the store.

**Stock decrement** is handled automatically by Clover when the order is paid,
provided items have `trackStock=true` and `stockCount` set in the Clover dashboard.

### Clover API Endpoints Used

| What                  | Method | Endpoint                                              |
|-----------------------|--------|-------------------------------------------------------|
| List inventory items  | GET    | `/v3/merchants/{mId}/items?expand=itemStock`          |
| Get single item       | GET    | `/v3/merchants/{mId}/items/{itemId}?expand=itemStock` |
| Create order          | POST   | `/v3/merchants/{mId}/orders`                          |
| Add line item         | POST   | `/v3/merchants/{mId}/orders/{orderId}/line_items`     |
| Pay for order         | POST   | `/v1/orders/{orderId}/pay` (Ecommerce API)            |
| Create refund         | POST   | `/v3/merchants/{mId}/refunds`                         |

The REST API (`/v3/...`) uses `CLOVER_BASE_URL`. The Ecommerce API (`/v1/...`) uses
`CLOVER_ECOMMERCE_URL`. Both authenticate with `CLOVER_PRIVATE_API_TOKEN` as a Bearer
token.

## CSS & Styling

- **Tailwind CSS v4** via standalone CLI binary (no Node.js/npm required)
- **Design**: Tan/beige background (#dfd9c8), dark foreground (#1b1a17), deep red primary (#912f2f), bright red accent (#fb2c36)
- **Dark mode**: Toggle with `.dark` class on `<html>`, stored in localStorage
- **Fonts**: Google Fonts — "IM Fell DW Pica" (serif), "Special Elite" (display), "Noto Sans"/"Nunito" (sans)
- Source: `static/input.css` — CSS custom properties for light/dark, `@theme inline` mapping to Tailwind colors, `.stamp` class, `.content` class, Clover `.field` styles
- Theme JS: `static/theme.js` — dark mode toggle + mobile menu (loaded in `<head>` to prevent FOUC)
- Output: `static/style.css` — generated, gitignored
- Binary: `bin/tailwindcss` — downloaded by `task tailwind:install`, gitignored
- `task css:build` — one-shot build; `task css:watch` — watch mode; `task dev` runs both watch + air
- Docker images download the `tailwindcss-linux-x64-musl` binary and build CSS in the image

### Known Limitations

- **Cart sessions expire after 30 days.** Session cookies have a 30-day MaxAge.
- **Stock requires Clover dashboard setup.** Items must have `trackStock=true` and
  an initial `stockCount` set in the Clover dashboard for auto-decrement to work.
