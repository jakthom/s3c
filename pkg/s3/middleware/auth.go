package s3middleware

import (
	"net/http"
	"strings"

	s3auth "github.com/jakthom/s3c/pkg/s3/auth"
	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

func AuthenticationMiddleware(authController s3auth.AuthController) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the request is authenticated
			authorizationHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authorizationHeader, "AWS4-HMAC-SHA256 ") {
				err := s3auth.AuthV4(w, r, authController, authorizationHeader)
				if err != nil {
					s3util.WriteError(w, r, err)
					return
				}
			} else {
				// Return access denied if the request doesn't use AWS auth V4
				// TODO -> add custom auth
				s3util.WriteError(w, r, s3error.AccessDeniedError(r))
			}
			next.ServeHTTP(w, r)
		})
	}
}
