# Security Guide

Security checklist for hosting on Digital Ocean with an admin panel.

## High Priority (Must Have)

### 1. HTTPS Everywhere
- Use Let's Encrypt (free) for TLS certificates
- Redirect all HTTP → HTTPS
- Use HSTS header
- Caddy handles this automatically

### 2. Admin Authentication
- Strong password hashing (bcrypt, argon2)
- Consider passwordless (magic link to email) or passkeys
- Rate limit login attempts
- Session timeout (e.g., 24 hours)
- Secure, HttpOnly, SameSite cookies

### 3. Server Hardening
```bash
# SSH key-only (disable password auth)
# Edit /etc/ssh/sshd_config:
PasswordAuthentication no
PermitRootLogin no

# Firewall: only expose ports 80, 443
ufw default deny incoming
ufw default allow outgoing
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 22/tcp  # SSH - consider limiting to your IP
ufw enable

# Automatic security updates
apt install unattended-upgrades
dpkg-reconfigure unattended-upgrades

# Fail2ban for brute force protection
apt install fail2ban
systemctl enable fail2ban
```

### 4. Secrets Management
- Never commit `.env` to git (add to `.gitignore`)
- Use environment variables or secrets manager
- Rotate keys if ever exposed
- Keep backups of secrets in secure location (password manager)

## Medium Priority (Should Have)

### 5. Application Security
- Input validation on all endpoints
- Parameterized queries (prevent SQL injection)
- CSRF tokens for forms
- Content-Security-Policy header
- Escape output (prevent XSS)

### 6. Minimal Attack Surface
- Don't expose admin panel publicly (IP whitelist or VPN)
- Or put admin on separate subdomain with extra auth
- Remove unused endpoints
- Keep dependencies updated (`go get -u ./...`)

### 7. Audit Logging
- Log all admin actions
- Log failed login attempts
- Log who checked in whom (for disputes)

## Lower Priority (Nice to Have)

- Two-factor authentication (TOTP)
- Database encryption at rest
- Backup encryption
- WAF (Web Application Firewall)
- Intrusion detection

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Digital Ocean                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │                   Firewall                        │  │
│  │              (only 80, 443, SSH)                  │  │
│  └───────────────────────────────────────────────────┘  │
│                         ↓                               │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Caddy (reverse proxy)                │  │
│  │         - Auto HTTPS via Let's Encrypt           │  │
│  │         - HSTS, security headers                 │  │
│  └───────────────────────────────────────────────────┘  │
│                         ↓                               │
│  ┌───────────────────────────────────────────────────┐  │
│  │                  Your Go App                      │  │
│  │         - Public: checkout, PWA                   │  │
│  │         - Admin: behind auth                      │  │
│  └───────────────────────────────────────────────────┘  │
│                         ↓                               │
│  ┌───────────────────────────────────────────────────┐  │
│  │               SQLite Database                     │  │
│  │            (file permissions 600)                 │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Admin Authentication Options

### Option 1: Magic Link (Recommended for simplicity)
- Admin enters email → gets login link → clicks → logged in
- No passwords to steal/crack
- Link expires in 15 minutes
- Requires email sending capability

### Option 2: Password + IP Whitelist
- Single admin password with bcrypt hashing
- Only allow admin access from specific IPs or VPN
- Even if password leaks, can't access remotely

### Option 3: Passkeys/WebAuthn
- Modern passwordless authentication
- Uses device biometrics or security keys
- Phishing resistant
- More complex to implement

## Caddy Security Headers

Example Caddyfile with security headers:

```
yourdomain.com {
    reverse_proxy localhost:8080

    header {
        # Security headers
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
        Referrer-Policy "strict-origin-when-cross-origin"
        Content-Security-Policy "default-src 'self'; script-src 'self' https://checkout.sandbox.dev.clover.com; frame-src https://checkout.sandbox.dev.clover.com"
    }
}
```

## Database Security (SQLite)

```bash
# Set proper file permissions
chmod 600 /path/to/database.db

# Regular backups
cp database.db backups/database-$(date +%Y%m%d).db

# Keep backups encrypted or in secure location
```

## Checklist Before Going Live

- [ ] HTTPS working (check with https://www.ssllabs.com/ssltest/)
- [ ] HTTP redirects to HTTPS
- [ ] Admin panel requires authentication
- [ ] SSH password auth disabled
- [ ] Firewall configured (only 80, 443, 22)
- [ ] Fail2ban installed and running
- [ ] `.env` not in git repository
- [ ] Automatic security updates enabled
- [ ] Database file has restricted permissions
- [ ] Backup strategy in place
- [ ] Audit logging enabled
