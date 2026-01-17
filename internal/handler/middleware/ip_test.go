package middleware

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTrustedIP(t *testing.T) {
	tests := []struct {
		name      string
		subnet    string
		wantError bool
	}{
		{"Valid IPv4 subnet", "192.168.1.0/24", false},
		{"Valid IPv6 subnet", "2001:db8::/32", false},
		{"Single IP", "10.0.0.1/32", false},
		{"Empty subnet", "", true},
		{"Invalid subnet", "invalid", true},
		{"Invalid CIDR", "192.168.1.0/33", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTrustedIP(tt.subnet)
			require.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestCheckIPMW_WithoutMiddleware(t *testing.T) {
	mw := CheckIPMW(nil)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	tests := []struct {
		name           string
		xRealIP        string
		expectedStatus int
	}{
		{"No header", "", http.StatusOK},
		{"Valid IP", "192.168.1.100", http.StatusOK},
		{"Invalid IP", "not-an-ip", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestCheckIPMW_WithMiddleware(t *testing.T) {
	trustedIP, err := NewTrustedIP("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	mw := CheckIPMW(trustedIP)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	tests := []struct {
		name           string
		xRealIP        string
		expectedStatus int
	}{
		{"Valid IP in subnet", "192.168.1.100", http.StatusOK},
		{"No header", "", http.StatusForbidden},
		{"IP outside subnet", "10.0.0.1", http.StatusForbidden},
		{"Invalid IP format", "test", http.StatusForbidden},
		{"Empty IP", "", http.StatusForbidden},
		{"different subnet", "192.168.2.100", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
