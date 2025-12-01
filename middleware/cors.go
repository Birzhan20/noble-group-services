package middleware

import (
	"net/http"
)

// CORSMiddleware adds CORS headers to the response.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Access-Control-Allow-Origin cannot handle multiple values comma-separated in strict browsers.
		// We need to check the Origin header and echo it back if it matches allowed list.
		origin := r.Header.Get("Origin")
		allowedOrigins := map[string]bool{
			"https://noble-group.vercel.app": true,
			"http://localhost:3000":          true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback or restrictive. For development, maybe allow all?
			// TS says "CORS allowed for https://noble-group.vercel.app and localhost:3000".
			// If origin is not sent (e.g. server-to-server), we might not set it, or set *.
			// But for browser, we should be specific.
			// Let's set to the first one as default if no match, or better yet, just leave it unset to block.
			// Or for simplicity in this task, if origin is missing, we don't care.
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
