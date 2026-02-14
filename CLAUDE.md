# LSD3 Project

Go web application for accepting payments for digital goods via Clover.

## Structure

```
cmd/server/main.go       - Entry point, router setup
internal/clover/         - Clover API client (payments, refunds)
internal/handlers/       - HTTP handlers (checkout, refunds)
pages/                   - HTML files served at /*
static/                  - JS/CSS assets served at /static/*
docs/                    - Documentation
  SECURITY.md            - Security guide for hosting
```

## Running

```bash
# Required environment variables (or use .env file)
export CLOVER_BASE_URL=https://sandbox.dev.clover.com  # or https://api.clover.com
export CLOVER_MERCHANT_ID=...
export CLOVER_PUBLIC_API_TOKEN=...   # For frontend SDK
export CLOVER_PRIVATE_API_TOKEN=...  # For backend API
export PORT=8080                     # optional

go run ./cmd/server
```

## API Endpoints

- `GET /api/config` - Returns public key for frontend SDK
- `POST /api/checkout` - Process payment with tokenized card
- `POST /api/refunds` - Process refund

## Payment Flow

1. Frontend loads Clover JS SDK with public key from `/api/config`
2. User fills form, Clover iframe captures card details
3. Frontend calls `clover.createToken()` to tokenize card
4. Frontend POSTs token + amount + customer info to `/api/checkout`
5. Backend charges token via Clover API

## Frontend

- `pages/index.html` - Checkout page
- `static/app.js` - Clover SDK integration
- `static/style.css` - Styling

Products are hardcoded in the HTML (digital goods, no inventory).
