package middlewares

import (
	"net"
	"net/http"
)

func IPFilter(trustedCIDR string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}

			_, ipnet, err := net.ParseCIDR(trustedCIDR)
			if err != nil {
				http.Error(w, "Invalid CIDR format", http.StatusInternalServerError)
				return
			}

			ip := net.ParseIP(clientIP)
			if ip == nil {
				http.Error(w, "Invalid IP address format", http.StatusForbidden)
				return
			}

			if !ipnet.Contains(ip) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
