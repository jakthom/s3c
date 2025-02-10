package s3notimplemented

import (
	"net/http"

	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

// NotImplementedHandler creates an endpoint that returns
// `NotImplementedError` responses. This can be used in places expecting a
// `HandlerFunc`, e.g. mux middleware.
func NotImplementedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s3util.WriteError(w, r, s3error.NotImplementedError(r))
	}
}
