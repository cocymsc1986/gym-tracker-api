package middleware

import (
	"log"
	"net/http"
	"strings"
)

type CORSMiddleware struct {
	allowedOrigins []string
}

func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

func (c *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS middleware hit: %s %s, Origin: %s", r.Method, r.URL.Path, r.Header.Get("Origin"))
		
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		if c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			log.Printf("Origin allowed: %s", origin)
		} else {
			log.Printf("Origin not allowed: %s", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// stripWWW removes a leading "www." from the host portion of an origin URL.
// e.g. "https://www.foo.com" -> "https://foo.com"
func stripWWW(origin string) string {
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(origin, scheme+"www.") {
			return scheme + origin[len(scheme+"www."):]
		}
	}
	return origin
}

func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	// Normalize www so "https://www.foo.com" matches an allowlist entry of "https://foo.com"
	normalized := stripWWW(origin)

	for _, allowedOrigin := range c.allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin || allowedOrigin == normalized {
			return true
		}
		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
