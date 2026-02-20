package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
)

const cookieName = "session_id"

// ContextKey is a typed key for request context values to avoid collisions.
type ContextKey string

const (
	// SessionIDKey is the context key for the session ID.
	SessionIDKey ContextKey = "sessionID"
	// CartCountKey is the context key for the cart item count.
	CartCountKey ContextKey = "cartCount"
)

// CartCounter returns the total item count for a session.
type CartCounter interface {
	CountItems(sessionID string) int
}

// validSessionID matches exactly 32 lowercase hex characters.
var validSessionID = regexp.MustCompile(`^[0-9a-f]{32}$`)

// Session injects the session ID and cart count into the request context.
func Session(counter CartCounter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := sessionFromCookie(r)
			if sessionID == "" {
				sessionID = newSessionID()
				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    sessionID,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   86400 * 30, // 30 days
				})
			}

			count := counter.CountItems(sessionID)

			ctx := r.Context()
			ctx = context.WithValue(ctx, SessionIDKey, sessionID)
			ctx = context.WithValue(ctx, CartCountKey, count)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sessionFromCookie(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	if !validSessionID.MatchString(c.Value) {
		return ""
	}
	return c.Value
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
