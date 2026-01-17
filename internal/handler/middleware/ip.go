package middleware

import (
	"net"
	"net/http"
)

type TrustedIP struct {
	trustedSubnet *net.IPNet
}

func NewTrustedIP(subnet string) (*TrustedIP, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	return &TrustedIP{trustedSubnet: ipnet}, nil
}

func CheckIPMW(cs *TrustedIP) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cs != nil {

				clientIP := r.Header.Get("X-Real-IP")
				if clientIP == "" {
					http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
					return
				}

				ip := net.ParseIP(clientIP)
				if ip == nil || !cs.trustedSubnet.Contains(ip) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
