package s3middleware

import (
	"net/http"

	"github.com/jakthom/s3c/pkg/header"
	"github.com/rs/zerolog/log"
)

// EtagMiddleware iterates over a request's headers and quotes unquoted Entity Tags headers.
// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
func EtagMiddleware(next http.Handler) http.Handler {
	etagHeaders := [3]string{"ETag", "If-Match", "If-None-Match"}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Interface("headers", r.Header).Msg("Headers before ETagMiddleware")
		for _, key := range etagHeaders {
			value := r.Header.Get(key)
			if value != "" {
				r.Header.Set(key, header.AddETagQuotes(value))
			}
		}
		next.ServeHTTP(w, r)
	})
}
