// Package sameorigin provides protection for state-changing HTTP routes.
package sameorigin

import (
	"net/http"
	"net/url"
	"strings"
)

// Require rejects cross-site requests using Origin or Referer checks.
func Require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requestMatches(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func requestMatches(r *http.Request) bool {
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return matchesRequestOrigin(origin, r)
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		return matchesRequestOrigin(referer, r)
	}
	return false
}

func matchesRequestOrigin(raw string, r *http.Request) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Scheme, RequestScheme(r)) && strings.EqualFold(u.Host, r.Host)
}

// RequestScheme returns the request scheme, honoring TLS termination forwarded by a trusted proxy.
// Deployments must ensure proxies strip untrusted X-Forwarded-Proto headers.
func RequestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme := strings.ToLower(strings.TrimSpace(strings.Split(forwarded, ",")[0]))
		if scheme == "http" || scheme == "https" {
			return scheme
		}
	}
	return "http"
}
