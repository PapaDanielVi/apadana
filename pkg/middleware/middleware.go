// Package middleware provides HTTP middleware for tenant identification.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	"net"
)

// Extractor extracts a tenant ID from an HTTP request.
type Extractor func(*http.Request) (string, error)

// FromHeader returns an Extractor that reads tenant ID from the named header.
func FromHeader(headerName string) Extractor {
	return func(r *http.Request) (string, error) {
		return r.Header.Get(headerName), nil
	}
}

// FromQuery returns an Extractor that reads tenant ID from the named query parameter.
func FromQuery(paramName string) Extractor {
	return func(r *http.Request) (string, error) {
		return r.URL.Query().Get(paramName), nil
	}
}

// FromSubdomain returns an Extractor that extracts tenant ID from the subdomain.
// e.g., "acme.example.com" → "acme".
func FromSubdomain() Extractor {
	return func(r *http.Request) (string, error) {
		host := r.Host
		if i := strings.Index(host, ":"); i != -1 {
			host = host[:i]
		}
		if net.ParseIP(host) != nil {
        	return "", errors.New("host is an ip address, cannot extract subdomain")
        }
		parts := strings.Split(host, ".")
		if len(parts) < 2 {
			return "", errors.New("no subdomain found")
		}
		return parts[0], nil
	}
}

// TenantMiddleware returns a middleware that extracts tenant ID via extractor
// and stores it in the request context.
func TenantMiddleware(extractor Extractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, err := extractor(r)
			if err != nil {
				tenantID = ""
			}
			if tenantID != "" {
				r = r.WithContext(tctx.WithTenantID(r.Context(), tenantID))
			}
			next.ServeHTTP(w, r)
		})
	}
}
