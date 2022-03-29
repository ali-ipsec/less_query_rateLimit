package limiters

import (
	"log"
	"net/http"
	"snapp/db"

	"golang.org/x/time/rate"
)

type KeyRateLimiter struct {
	// limiter *rate.Limiter
	keys map[string]*rate.Limiter
	r    rate.Limit
	b    int
}

func NewKeyRateLimiter(r rate.Limit, b int) *KeyRateLimiter {
	return &KeyRateLimiter{
		keys: make(map[string]*rate.Limiter),
		r:    r,
		b:    b,
	}
}
func (i *KeyRateLimiter) AddKey(key string) *rate.Limiter {
	limiter := rate.NewLimiter(i.r, i.b)
	i.keys[key] = limiter
	return limiter
}

func (i *KeyRateLimiter) getLimiter(key string) *rate.Limiter {
	limiter, exists := i.keys[key]
	if !exists {
		return i.AddKey(key)
	} else {

		return limiter
	}

}
func ByAppKey(next http.Handler, refillRate rate.Limit, tokenBucketSize int) http.Handler {
	// TODO: Implement
	KeyRateLimiter := NewKeyRateLimiter(refillRate, tokenBucketSize)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := r.Header.Get("X-App-Key")
		limiter := KeyRateLimiter.getLimiter(k)

		rows, err := db.GetConnection().Query("select id from app_keys where key=?", k)
		if err != nil {
			log.Print(err)
		}

		defer rows.Close()
		var id int
		for rows.Next() {
			err := rows.Scan(&id)
			if err != nil {
				log.Fatal(err)
			} else {
				if !limiter.Allow() {
					data := []byte(`{"error": "too many requests"}`)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write(data)
					return
				}
			}
			next.ServeHTTP(w, r)

		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

	})

}
