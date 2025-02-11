package s3handler

import (
	"net/http"

	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
	"github.com/rs/zerolog/log"
)

func MethodNotAllowedHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Error().Str("method", r.Method).Str("path", r.URL.Path).Msg("method not allowed")
		s3util.WriteError(w, r, s3error.MethodNotAllowedError(r))
	})
}
