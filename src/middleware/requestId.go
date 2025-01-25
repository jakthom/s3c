package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const REQUEST_ID = "requestId"

// RequestIDMiddleware adds a request ID to every request.
func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := uuid.New()
		vars["requestID"] = id.String()
		next.ServeHTTP(w, r)
	})
}
