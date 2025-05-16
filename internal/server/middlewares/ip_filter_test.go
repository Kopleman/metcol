package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPFilter(t *testing.T) {
	tests := []struct {
		name           string
		trustedCIDR    string
		clientIP       string
		expectedStatus int
	}{
		{
			name:           "Valid IP in trusted subnet",
			trustedCIDR:    "192.168.1.0/24",
			clientIP:       "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet",
			trustedCIDR:    "192.168.1.0/24",
			clientIP:       "192.168.2.100",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing X-Real-IP header",
			trustedCIDR:    "192.168.1.0/24",
			clientIP:       "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid IP format",
			trustedCIDR:    "192.168.1.0/24",
			clientIP:       "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv6 in trusted subnet",
			trustedCIDR:    "2001:db8::/32",
			clientIP:       "2001:db8::1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 not in trusted subnet",
			trustedCIDR:    "2001:db8::/32",
			clientIP:       "2001:db9::1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := IPFilter(tt.trustedCIDR)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}

			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestIPFilterInvalidCIDR(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := IPFilter("invalid-cidr")
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
