# LSD3 — Remaining Todos

## Content
- [x] Fill in real chapter name (currently "Clamper Chapter" placeholder)
- [x] Write actual history page content
- [x] Write events page content
- [x] Write members page content

## Features
- [x] Clear cart on successful payment in the frontend (redirect or show empty cart)
- [x] Display stock count or "low stock" on shop page
- [x] Add email receipt via Clover `receipt_email` field on payment
- [x] Add error handling on shop page when Clover API is unreachable (show cached or message)

## Infrastructure
- [x] Persist carts (e.g., SQLite or file) so they survive restarts
- [x] Add HTTPS/TLS termination docs or Caddy config
- [x] Set up production Clover credentials and switch from sandbox
- [x] Add health check endpoint
