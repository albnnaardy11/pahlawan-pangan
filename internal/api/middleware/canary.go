package middleware

import (
	"math/rand/v2"
	"net/http"
)

// CanarySplitter splits traffic between two versions.
// Essential for Blast Radius Control during 287M user rollout.
func CanarySplitter(canaryPercentage int, canaryHeader string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Probability check
			if rand.IntN(100) < canaryPercentage {
				w.Header().Set("X-Traffic-Type", "CANARY")
				w.Header().Set("X-App-Version", canaryHeader)
			} else {
				w.Header().Set("X-Traffic-Type", "STABLE")
				w.Header().Set("X-App-Version", "v1.0.0")
			}
			next.ServeHTTP(w, r)
		})
	}
}
