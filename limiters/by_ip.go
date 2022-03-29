package limiters

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	// limiter *rate.Limiter
	ips map[string]*rate.Limiter
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	fmt.Println("NewIPRateLimiter", r, b)
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	fmt.Println("addIP", i.r, i.b, ip)
	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter
	return limiter
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	limiter, exists := i.ips[ip]
	fmt.Println("existance", ip, exists)
	if !exists {
		return i.AddIP(ip)
	} else {

		return limiter
	}

}
func ByIp(next http.Handler, refillRate rate.Limit, tokenBucketSize int) http.Handler {
	// TODO: Implement
	ipLimiter := NewIPRateLimiter(refillRate, tokenBucketSize)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := ipLimiter.getLimiter(r.RemoteAddr)
		fmt.Println("allow", limiter.Allow(), !limiter.Allow())
		if !limiter.Allow() {
			fmt.Println("KKKKKKKKKKKKKKKKKKKKKKKKKK")
			data := []byte(`{"error": "too many requests"}`)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write(data)
		}

		next.ServeHTTP(w, r)
	})

}
