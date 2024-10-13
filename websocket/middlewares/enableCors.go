package middlewares

import (
	"net/http"
)

// CORS middleware to add the necessary headers
func EnableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request
		if req.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(writer, req)
	})
}
