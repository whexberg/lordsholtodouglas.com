package middleware

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

// CSRF returns middleware that rejects cross-origin state-changing requests.
// It checks the Origin header (preferred) or Referer header against the
// request's Host. Safe methods (GET, HEAD, OPTIONS) are always allowed.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			if !sameHost(origin, r.Host) {
				log.Printf("csrf: rejected cross-origin %s %s (origin=%s host=%s)", r.Method, r.URL.Path, origin, r.Host)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// No Origin header — fall back to Referer.
		referer := r.Header.Get("Referer")
		if referer != "" {
			if !sameHost(referer, r.Host) {
				log.Printf("csrf: rejected cross-origin %s %s (referer=%s host=%s)", r.Method, r.URL.Path, referer, r.Host)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Neither Origin nor Referer present. This can happen with direct
		// API calls (curl, Postman) or privacy-stripping browsers. Reject
		// form POSTs (browsers always send at least one header) but allow
		// JSON API calls which are protected by CORS preflight.
		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("csrf: rejected %s %s (no origin or referer, content-type=%s)", r.Method, r.URL.Path, ct)
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

// sameHost checks whether the given URL string's host matches the expected host.
func sameHost(rawURL string, host string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return stripPort(u.Host) == stripPort(host)
}

func stripPort(hostport string) string {
	host, _, found := strings.Cut(hostport, ":")
	if found {
		return host
	}
	return hostport
}
